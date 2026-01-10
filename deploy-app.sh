#!/bin/bash
set -e

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}Deploy Application to EC2${NC}"
echo -e "${GREEN}========================================${NC}\n"

# Check if deployment-info.txt exists
if [ ! -f "deployment-info.txt" ]; then
    echo -e "${RED}Error: deployment-info.txt not found${NC}"
    echo -e "${YELLOW}Please run ./demo-deploy.sh first${NC}"
    exit 1
fi

# Extract instance IP from deployment-info.txt
PUBLIC_IP=$(grep "Public IP:" deployment-info.txt | awk '{print $3}')
KEY_FILE=$(grep "SSH Key:" deployment-info.txt | awk '{print $3}')

if [ -z "$PUBLIC_IP" ] || [ -z "$KEY_FILE" ]; then
    echo -e "${RED}Error: Could not read deployment info${NC}"
    exit 1
fi

echo -e "${YELLOW}Target Instance: $PUBLIC_IP${NC}"
echo -e "${YELLOW}SSH Key: $KEY_FILE${NC}\n"

# Wait for instance to be ready
echo -e "${YELLOW}Checking if instance is ready...${NC}"
for i in {1..10}; do
    if ssh -o StrictHostKeyChecking=no -o ConnectTimeout=5 -i "$KEY_FILE" ec2-user@$PUBLIC_IP "cat /tmp/setup-status 2>/dev/null" | grep -q "ready"; then
        echo -e "${GREEN}✓ Instance is ready${NC}\n"
        break
    fi
    if [ $i -eq 10 ]; then
        echo -e "${YELLOW}Instance may still be initializing. Continuing anyway...${NC}\n"
    fi
    echo "Waiting... ($i/10)"
    sleep 10
done

# Create remote deployment script
cat > /tmp/remote-deploy.sh << 'REMOTESCRIPT'
#!/bin/bash
set -e

cd /opt/ticketing-system

# Check if repo exists, if not, create placeholder
if [ ! -d ".git" ]; then
    echo "Setting up application directory..."
    
    # Create a simple main.go for testing
    cat > main.go << 'EOF'
package main

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "os"
    "time"
)

func main() {
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }

    http.HandleFunc("/", homeHandler)
    http.HandleFunc("/health", healthHandler)
    http.HandleFunc("/api/status", statusHandler)

    fmt.Printf("Server starting on port %s...\n", port)
    fmt.Printf("Access at: http://localhost:%s\n", port)
    log.Fatal(http.ListenAndServe(":"+port, nil))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "text/html")
    html := `
    <!DOCTYPE html>
    <html>
    <head>
        <title>Ticketing System Demo</title>
        <style>
            body {
                font-family: Arial, sans-serif;
                max-width: 800px;
                margin: 50px auto;
                padding: 20px;
                background: #f5f5f5;
            }
            .container {
                background: white;
                padding: 30px;
                border-radius: 10px;
                box-shadow: 0 2px 10px rgba(0,0,0,0.1);
            }
            h1 { color: #333; }
            .status { color: #28a745; font-weight: bold; }
            .endpoint {
                background: #f8f9fa;
                padding: 15px;
                margin: 10px 0;
                border-left: 4px solid #007bff;
            }
            code {
                background: #e9ecef;
                padding: 2px 6px;
                border-radius: 3px;
            }
        </style>
    </head>
    <body>
        <div class="container">
            <h1>🎫 Ticketing System - Demo Server</h1>
            <p class="status">✓ Server is running!</p>
            
            <h2>Available Endpoints:</h2>
            
            <div class="endpoint">
                <strong>GET /health</strong><br>
                Health check endpoint<br>
                <code>curl http://` + r.Host + `/health</code>
            </div>
            
            <div class="endpoint">
                <strong>GET /api/status</strong><br>
                System status endpoint<br>
                <code>curl http://` + r.Host + `/api/status</code>
            </div>
            
            <h2>Next Steps:</h2>
            <ol>
                <li>Clone your ticketing system repository</li>
                <li>Build the application: <code>go build -o api-server ./cmd/api-server</code></li>
                <li>Run migrations: <code>cd migrations && go run main.go</code></li>
                <li>Start the server: <code>./api-server</code></li>
            </ol>
            
            <h2>Database Info:</h2>
            <ul>
                <li>PostgreSQL: localhost:5432</li>
                <li>Database: ticketing_db</li>
                <li>User: ticketing_user</li>
                <li>Redis: localhost:6379</li>
            </ul>
        </div>
    </body>
    </html>
    `
    fmt.Fprint(w, html)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "status": "healthy",
        "time":   time.Now().Format(time.RFC3339),
    })
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "application": "Ticketing System",
        "version":     "demo-1.0",
        "status":      "running",
        "timestamp":   time.Now().Format(time.RFC3339),
        "message":     "Demo server is operational",
    })
}
EOF

    # Create go.mod
    cat > go.mod << 'EOF'
module demo-server

go 1.22
EOF

fi

# Build the application
echo "Building application..."
/usr/local/go/bin/go build -o demo-server main.go

# Create systemd service
sudo tee /etc/systemd/system/ticketing-demo.service > /dev/null << 'EOF'
[Unit]
Description=Ticketing System Demo Server
After=network.target docker.service
Requires=docker.service

[Service]
Type=simple
User=ec2-user
WorkingDirectory=/opt/ticketing-system
ExecStart=/opt/ticketing-system/demo-server
Restart=always
RestartSec=5s
Environment="PORT=8080"
Environment="DB_HOST=localhost"
Environment="DB_PORT=5432"
Environment="DB_NAME=ticketing_db"
Environment="DB_USER=ticketing_user"
Environment="DB_PASSWORD=ChangeMe123!"
Environment="REDIS_HOST=localhost"
Environment="REDIS_PORT=6379"

[Install]
WantedBy=multi-user.target
EOF

# Reload systemd and start service
sudo systemctl daemon-reload
sudo systemctl enable ticketing-demo
sudo systemctl restart ticketing-demo

echo "Application deployed successfully!"
REMOTESCRIPT

# Copy and execute the script
echo -e "${YELLOW}Deploying application to server...${NC}"
scp -o StrictHostKeyChecking=no -i "$KEY_FILE" /tmp/remote-deploy.sh ec2-user@$PUBLIC_IP:/tmp/
ssh -o StrictHostKeyChecking=no -i "$KEY_FILE" ec2-user@$PUBLIC_IP "bash /tmp/remote-deploy.sh"

echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  Deployment Complete!${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo -e "Your demo server is now running at:"
echo -e "${GREEN}http://$PUBLIC_IP:8080${NC}"
echo ""
echo -e "Try these commands:"
echo -e "${YELLOW}curl http://$PUBLIC_IP:8080/health${NC}"
echo -e "${YELLOW}curl http://$PUBLIC_IP:8080/api/status${NC}"
echo ""
echo -e "Or open in browser: ${GREEN}http://$PUBLIC_IP:8080${NC}"
echo ""
echo -e "To view logs:"
echo -e "${YELLOW}ssh -i $KEY_FILE ec2-user@$PUBLIC_IP 'sudo journalctl -u ticketing-demo -f'${NC}"
echo ""
echo -e "To deploy your actual application:"
echo -e "1. SSH: ${YELLOW}ssh -i $KEY_FILE ec2-user@$PUBLIC_IP${NC}"
echo -e "2. Clone: ${YELLOW}cd /opt/ticketing-system && git clone YOUR_REPO .${NC}"
echo -e "3. Build: ${YELLOW}go build -o api-server ./cmd/api-server${NC}"
echo -e "4. Update service to use your binary"
echo ""
