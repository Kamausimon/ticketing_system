# AWS SAA-C03 Study Guide - Integrated with Your Project

## Using Your Ticketing System to Learn AWS

You've already implemented many AWS concepts! Here's how your project maps to SAA-C03 topics:

---

## Your Project → AWS SAA-C03 Mapping

### ✅ Already Implemented in Your Project

#### 1. **Compute (EC2)**
**Your Experience:**
- Deployed ticketing system on EC2 instances
- Configured security groups
- Used user data scripts for initialization

**SAA-C03 Topics to Study:**
- [ ] Instance types and use cases (compute-optimized, memory-optimized)
- [ ] Placement groups
- [ ] Elastic IPs
- [ ] EC2 pricing models (On-Demand, Reserved, Spot)
- [ ] Auto Scaling Groups (ASG)
- [ ] Launch templates vs Launch configurations

**Practice Exercise:**
```
Create an Auto Scaling Group for your ticketing system:
1. Create launch template with your app
2. Configure ASG with min/desired/max instances
3. Add scaling policies based on CPU utilization
4. Test scale-out and scale-in events
```

#### 2. **VPC & Networking**
**Your Experience:**
- Created VPC with CIDR blocks
- Configured public/private subnets
- Set up Internet Gateway and NAT Gateway
- Configured Route Tables
- Implemented Security Groups

**SAA-C03 Topics to Study:**
- [ ] CIDR notation and IP addressing
- [ ] VPC Peering vs Transit Gateway
- [ ] VPC Endpoints (Gateway vs Interface)
- [ ] Network ACLs vs Security Groups
- [ ] VPN and Direct Connect
- [ ] Route 53 for DNS

**Practice Exercise:**
```
Extend your VPC design:
1. Add private subnet for databases
2. Create VPC endpoint for S3 access (no internet)
3. Set up Network ACLs for additional security
4. Configure VPC Flow Logs to CloudWatch
```

#### 3. **Database (RDS)**
**Your Experience:**
- Deployed PostgreSQL on RDS
- Configured in private subnet
- Set up security groups for database access

**SAA-C03 Topics to Study:**
- [ ] Multi-AZ vs Read Replicas
- [ ] Backup and snapshot strategies
- [ ] RDS Proxy for connection pooling
- [ ] Aurora vs standard RDS
- [ ] DynamoDB for NoSQL use cases
- [ ] Database migration services (DMS)

**Practice Exercise:**
```
Enhance your database setup:
1. Enable Multi-AZ deployment
2. Create read replica for reporting queries
3. Configure automated backups with 7-day retention
4. Set up CloudWatch alarms for database metrics
5. Test failover scenario
```

#### 4. **Caching (ElastiCache)**
**Your Experience:**
- Deployed Redis via ElastiCache
- Implemented caching strategy in your app

**SAA-C03 Topics to Study:**
- [ ] Redis vs Memcached comparison
- [ ] Cluster mode enabled vs disabled
- [ ] Backup and restore strategies
- [ ] Redis AUTH and encryption

**Practice Exercise:**
```
Optimize your caching:
1. Enable cluster mode for horizontal scaling
2. Configure Redis AUTH for security
3. Set up CloudWatch metrics for cache hit ratio
4. Implement cache warming strategy
```

#### 5. **Security & IAM**
**Your Experience:**
- Created IAM roles for EC2
- Configured Security Groups
- Implemented WAF rules

**SAA-C03 Topics to Study:**
- [ ] IAM policies (managed vs inline)
- [ ] IAM roles vs users vs groups
- [ ] Cross-account access
- [ ] S3 bucket policies
- [ ] KMS for encryption
- [ ] Secrets Manager vs Systems Manager Parameter Store
- [ ] AWS Organizations and SCPs

**Practice Exercise:**
```
Enhance security:
1. Store DB password in Secrets Manager (not environment variables)
2. Create least-privilege IAM policy for EC2
3. Enable encryption at rest for RDS using KMS
4. Set up AWS Config for compliance checking
5. Implement S3 bucket policy for static assets
```

#### 6. **Monitoring (CloudWatch)**
**Your Experience:**
- Set up CloudWatch alarms
- Configured metrics collection

**SAA-C03 Topics to Study:**
- [ ] CloudWatch Logs, Metrics, and Alarms
- [ ] CloudWatch Events vs EventBridge
- [ ] X-Ray for distributed tracing
- [ ] Custom metrics
- [ ] CloudWatch Logs Insights for querying

**Practice Exercise:**
```
Complete monitoring stack:
1. Create custom metrics (ticket sales per hour)
2. Set up CloudWatch dashboard
3. Configure SNS topic for alarm notifications
4. Implement X-Ray tracing in your Go app
5. Create log metric filters for error tracking
```

---

## SAA-C03 Topics NOT in Your Project (Must Study)

### 1. **S3 (Simple Storage Service)**
**What to Learn:**
- Storage classes (Standard, IA, Glacier)
- Lifecycle policies
- Versioning and MFA delete
- Cross-region replication
- S3 Transfer Acceleration
- Pre-signed URLs

**Add to Your Project:**
```go
// Store ticket PDFs and event images in S3
// Practice:
- Upload event images to S3
- Generate pre-signed URLs for ticket PDFs
- Implement lifecycle policy (move old tickets to Glacier)
- Enable versioning for event images
```

### 2. **Lambda & Serverless**
**What to Learn:**
- Lambda function configuration
- Event triggers (S3, API Gateway, CloudWatch)
- Lambda@Edge
- Step Functions for orchestration
- API Gateway integration

**Add to Your Project:**
```
Ideas to practice Lambda:
- Resize uploaded event images automatically
- Generate PDF tickets asynchronously
- Send email notifications via Lambda
- Create daily sales report Lambda (EventBridge schedule)
```

### 3. **Load Balancing**
**What to Learn:**
- ALB vs NLB vs CLB
- Target groups and health checks
- Sticky sessions
- Cross-zone load balancing
- ALB routing rules (path-based, host-based)

**Add to Your Project:**
```
Enhance your ALB setup:
1. Configure path-based routing (/api/* vs /admin/*)
2. Set up multiple target groups
3. Implement health check endpoints
4. Add HTTPS with ACM certificate
5. Configure WAF rules on ALB
```

### 4. **Container Services (ECS, EKS)**
**What to Learn:**
- ECS vs EKS vs Fargate
- Task definitions
- Service discovery
- Container insights

**Add to Your Project:**
```
Containerize your app:
1. Create Dockerfile (already done!)
2. Deploy to ECS Fargate
3. Configure service auto-scaling
4. Set up ALB with ECS
```

### 5. **High Availability & Disaster Recovery**
**What to Learn:**
- Multi-AZ vs Multi-Region
- RPO and RTO concepts
- Backup strategies
- Route 53 health checks and failover

**Practice Scenarios:**
```
Design solutions for:
1. 99.99% uptime requirement
2. RPO of 1 hour, RTO of 4 hours
3. Active-active multi-region setup
4. Database failover strategy
```

### 6. **Cost Optimization**
**What to Learn:**
- Pricing models comparison
- Cost allocation tags
- Reserved instances vs Savings Plans
- AWS Cost Explorer
- Trusted Advisor

**Practice:**
```
Optimize your demo costs:
1. Use t3.micro instead of t3.medium
2. Set up billing alerts
3. Tag all resources
4. Review Trusted Advisor recommendations
5. Consider Spot instances for non-critical tasks
```

---

## SAA-C03 Exam Domains & Your Coverage

### Domain 1: Design Secure Architectures (30%)
**You've Covered:**
✅ Security Groups and NACLs
✅ IAM roles for EC2
✅ Data encryption (RDS)
✅ VPC security

**Still Need:**
- [ ] KMS key management
- [ ] Secrets Manager
- [ ] AWS Shield and WAF (deepen knowledge)
- [ ] GuardDuty for threat detection
- [ ] S3 bucket policies and ACLs

### Domain 2: Design Resilient Architectures (26%)
**You've Covered:**
✅ Multi-AZ RDS
✅ Auto Scaling Groups
✅ Load Balancing

**Still Need:**
- [ ] Multi-region architectures
- [ ] Disaster recovery strategies
- [ ] Route 53 failover routing
- [ ] S3 cross-region replication
- [ ] Backup and restore procedures

### Domain 3: Design High-Performing Architectures (24%)
**You've Covered:**
✅ ElastiCache (Redis)
✅ RDS Read Replicas
✅ CloudWatch monitoring

**Still Need:**
- [ ] CloudFront CDN
- [ ] S3 Transfer Acceleration
- [ ] Database sharding strategies
- [ ] Caching strategies (CloudFront, API Gateway)
- [ ] Performance monitoring and optimization

### Domain 4: Design Cost-Optimized Architectures (20%)
**You've Covered:**
✅ Basic cost awareness

**Still Need:**
- [ ] Reserved Instances vs Savings Plans
- [ ] S3 storage classes
- [ ] Cost allocation tags
- [ ] AWS Budgets
- [ ] Instance right-sizing

---

## 30-Day SAA-C03 Study Plan (While Building)

### Week 1: Strengthen Your Foundations
**Days 1-2: VPC Deep Dive**
- [ ] Study: CIDR, subnets, routing tables
- [ ] Practice: Design 3-tier VPC architecture
- [ ] Document your current VPC setup
- [ ] Resources: AWS VPC Documentation, VPC Whitepaper

**Days 3-4: Security & IAM**
- [ ] Study: IAM policies, roles, best practices
- [ ] Practice: Create least-privilege policies
- [ ] Migrate to Secrets Manager
- [ ] Resources: IAM Best Practices Guide

**Days 5-7: Compute (EC2)**
- [ ] Study: Instance types, pricing, Auto Scaling
- [ ] Practice: Create ASG for your app
- [ ] Implement launch template
- [ ] Resources: EC2 FAQs, Auto Scaling Documentation

### Week 2: Storage & Databases
**Days 8-9: S3**
- [ ] Study: Storage classes, lifecycle, versioning
- [ ] Practice: Add S3 to your project
- [ ] Store event images and ticket PDFs
- [ ] Resources: S3 Developer Guide

**Days 10-11: RDS & Databases**
- [ ] Study: Multi-AZ, Read Replicas, Aurora
- [ ] Practice: Enable Multi-AZ, create read replica
- [ ] Test failover scenarios
- [ ] Resources: RDS Best Practices

**Days 12-14: Caching & Performance**
- [ ] Study: ElastiCache, CloudFront
- [ ] Practice: Optimize your Redis usage
- [ ] Add CloudFront for static content
- [ ] Resources: Caching Best Practices

### Week 3: Advanced Services
**Days 15-16: Load Balancing**
- [ ] Study: ALB, NLB, routing rules
- [ ] Practice: Configure advanced ALB rules
- [ ] Add HTTPS with ACM
- [ ] Resources: ELB Documentation

**Days 17-18: Serverless (Lambda)**
- [ ] Study: Lambda, API Gateway, Step Functions
- [ ] Practice: Add Lambda for image processing
- [ ] Create async ticket generation
- [ ] Resources: Lambda Developer Guide

**Days 19-21: Containers (ECS/EKS)**
- [ ] Study: ECS, Fargate, EKS
- [ ] Practice: Deploy your app to ECS
- [ ] Configure service discovery
- [ ] Resources: ECS Workshop

### Week 4: Exam Prep
**Days 22-23: High Availability & DR**
- [ ] Study: Multi-region, failover, backup
- [ ] Practice: Design HA architecture
- [ ] Create DR plan
- [ ] Resources: Well-Architected Framework

**Days 24-25: Cost Optimization**
- [ ] Study: Pricing models, cost tools
- [ ] Practice: Optimize your demo costs
- [ ] Create cost report
- [ ] Resources: Cost Optimization Pillar

**Days 26-28: Practice Exams**
- [ ] Tutorials Dojo Practice Exams (6 sets)
- [ ] Review incorrect answers
- [ ] Document weak areas
- [ ] Resources: Official AWS Practice Exam

**Days 29-30: Final Review**
- [ ] Review AWS Cheat Sheets
- [ ] Revisit weak topics
- [ ] Take final practice exam
- [ ] Schedule certification exam

---

## Study Resources (Prioritized)

### Must-Have (Free)
1. **AWS Documentation** - Official docs for each service
2. **AWS Well-Architected Framework** - Core reading
3. **AWS Whitepapers** - Security, storage, disaster recovery
4. **AWS Skill Builder** - Free digital training
5. **Your Project** - Hands-on practice!

### Highly Recommended (Paid)
1. **Tutorials Dojo Practice Exams** - $15-20 (BEST for exam prep)
2. **Stephane Maarek's Course (Udemy)** - $15 (wait for sale)
3. **Adrian Cantrill's Course** - $40 (very thorough)

### Supplementary
- AWS re:Invent videos on YouTube
- r/AWSCertifications subreddit
- AWS Architecture Center case studies
- freeCodeCamp AWS tutorials

---

## Weekly Study Schedule

**Monday-Friday (2 hours/day):**
- Morning (1 hour): Video course/reading
- Evening (1 hour): Hands-on practice in your project

**Saturday (4 hours):**
- Deep dive into one service
- Build feature in your project using that service
- Document what you learned

**Sunday (2 hours):**
- Review week's topics
- Practice exam questions
- Update study notes

**Total: 16 hours/week**

---

## Hands-On Projects to Enhance Learning

### Project 1: Multi-Region Deployment
```
Deploy your ticketing system in two regions:
- Primary: us-east-1
- Secondary: eu-west-1
- Use Route 53 for failover
- Replicate database using DMS
- Test failover scenario
```

### Project 2: Serverless Ticket Processing
```
Add Lambda functions for:
- Image resizing (S3 trigger)
- Email notifications (SQS/SNS)
- Daily sales reports (EventBridge)
- Ticket PDF generation
```

### Project 3: Cost-Optimized Architecture
```
Redesign for minimal cost:
- Use Spot instances for workers
- Move old data to S3 Glacier
- Implement aggressive caching
- Use RDS Read Replicas for reports
- Target: <$50/month for production-like setup
```

### Project 4: High Availability Setup
```
Build 99.99% uptime architecture:
- Multi-AZ in all services
- Auto Scaling with predictive scaling
- Health checks and automated recovery
- CloudWatch dashboards and alarms
- Document RTO and RPO
```

---

## Exam Day Checklist

### One Week Before:
- [ ] Score 80%+ on practice exams consistently
- [ ] Review all AWS service limits and quotas
- [ ] Memorize common CIDR blocks
- [ ] Know S3 storage classes by heart
- [ ] Understand Well-Architected pillars

### Day Before:
- [ ] Light review only (no new topics)
- [ ] Review your cheat sheets
- [ ] Get good sleep
- [ ] Prepare ID and confirmation

### Exam Day:
- [ ] Arrive 15 minutes early (or start on time if online)
- [ ] Flag uncertain questions for review
- [ ] Manage time: ~1.5 minutes per question
- [ ] Read questions carefully (key words: most cost-effective, least operational overhead)
- [ ] Trust your hands-on experience!

---

## Post-Certification

After passing SAA-C03:
1. Update LinkedIn, resume
2. Add badge to email signature
3. Share success story on LinkedIn
4. Consider next cert: Developer, SysOps, or Solutions Architect Professional

---

## Your Competitive Advantage

Most SAA-C03 candidates study theory without practice. You have:
✅ Real production architecture experience
✅ Hands-on AWS deployment
✅ Understanding of real-world trade-offs
✅ Working knowledge of security and performance
✅ Cost optimization experience

**You're ahead of 70% of candidates just by having your project!**

---

**Start today! Deploy on Railway this week, then enhance with AWS features while studying! 🚀**

Your ticketing system = Your AWS playground = Your certification success! 🎯
