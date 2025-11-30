# Concurrency Fixes - Deployment Checklist

## Pre-Deployment Review

### Code Review
- [x] Review `internal/orders/create.go` changes
  - [x] Pessimistic locking implemented (SELECT FOR UPDATE)
  - [x] Re-check after lock acquisition
  - [x] Optimistic locking with version field
  - [x] Proper error handling
  - [x] Transaction rollback on failures

- [x] Review `internal/settlement/calculate.go` changes
  - [x] All calculations in single transaction
  - [x] Consistent data snapshot
  - [x] Proper error handling

- [x] Review `internal/models/ticketsClasses.go` changes
  - [x] Version field added with default value
  - [x] Proper GORM tags

- [x] Code compiles without errors
- [x] No linting issues
- [x] Documentation complete

### Testing
- [ ] Run unit tests
  ```bash
  cd internal/orders
  go test -v -run TestConcurrent
  ```

- [ ] Run benchmarks
  ```bash
  go test -bench=BenchmarkConcurrent -benchmem
  ```

- [ ] Manual testing in dev environment
  - [ ] Single ticket purchase (baseline)
  - [ ] Concurrent purchases (2 users, 1 ticket)
  - [ ] High load test (10+ concurrent requests)
  - [ ] Settlement calculation test

### Database Preparation
- [ ] Backup production database
  ```bash
  pg_dump -U postgres ticketing_system > backup_before_concurrency_fix.sql
  ```

- [ ] Test migration on staging database
  ```bash
  cd migrations && go run main.go
  ```

- [ ] Verify migration success
  ```sql
  SELECT column_name, data_type, column_default 
  FROM information_schema.columns 
  WHERE table_name = 'ticket_classes' AND column_name = 'version';
  ```

- [ ] Rollback test on staging
  ```sql
  ALTER TABLE ticket_classes DROP COLUMN version;
  ```

## Deployment Steps

### 1. Database Migration (5 minutes)
- [ ] Run GORM migration
  ```bash
  cd migrations
  go run main.go
  ```

- [ ] Verify migration
  ```sql
  -- Check version column exists
  SELECT version FROM ticket_classes LIMIT 5;
  
  -- Check all rows have version = 0
  SELECT COUNT(*) FROM ticket_classes WHERE version != 0;
  -- Should return 0
  
  -- Check index exists
  \d ticket_classes
  ```

- [ ] Record migration time: _______

### 2. Application Deployment (10 minutes)
- [ ] Build new binary
  ```bash
  go build -o bin/api-server ./cmd/api-server
  ```

- [ ] Stop current application
  ```bash
  sudo systemctl stop ticketing-api
  ```

- [ ] Deploy new binary
  ```bash
  sudo cp bin/api-server /opt/ticketing/api-server
  sudo chmod +x /opt/ticketing/api-server
  ```

- [ ] Start application
  ```bash
  sudo systemctl start ticketing-api
  ```

- [ ] Check application logs
  ```bash
  sudo journalctl -u ticketing-api -f
  ```

- [ ] Verify health endpoint
  ```bash
  curl http://localhost:8080/health
  ```

### 3. Smoke Tests (5 minutes)
- [ ] Test single ticket purchase
  ```bash
  curl -X POST http://localhost:8080/api/orders \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"event_id": 1, "items": [{"ticket_class_id": 1, "quantity": 1}], ...}'
  ```

- [ ] Test settlement calculation
  ```bash
  curl -X POST http://localhost:8080/api/settlements/calculate \
    -H "Authorization: Bearer $ADMIN_TOKEN" \
    -d '{"event_id": 1}'
  ```

- [ ] Check database state
  ```sql
  -- Verify version increments after purchase
  SELECT id, quantity_sold, version FROM ticket_classes WHERE id = 1;
  ```

- [ ] Check application metrics
  - [ ] Error rate: < 0.1%
  - [ ] Response time: < 200ms
  - [ ] Lock wait time: < 50ms

## Post-Deployment Monitoring

### First Hour
- [ ] Monitor error logs
  ```bash
  tail -f /var/log/ticketing/error.log | grep -i "ticket inventory"
  ```

- [ ] Monitor lock metrics
  ```sql
  SELECT wait_event_type, wait_event, count(*) 
  FROM pg_stat_activity 
  WHERE state = 'active' 
  GROUP BY wait_event_type, wait_event;
  ```

- [ ] Check for version conflicts
  ```bash
  grep "ticket inventory changed during checkout" /var/log/ticketing/app.log | wc -l
  ```

- [ ] Monitor transaction duration
  ```sql
  SELECT 
    query,
    state,
    now() - query_start as duration
  FROM pg_stat_activity
  WHERE query LIKE '%ticket_classes%'
  ORDER BY duration DESC
  LIMIT 10;
  ```

### First 24 Hours
- [ ] Daily metrics report
  - [ ] Total orders: _______
  - [ ] Failed orders: _______
  - [ ] Version conflicts: _______
  - [ ] Average lock wait: _______ms
  - [ ] 95th percentile latency: _______ms

- [ ] Check for deadlocks
  ```sql
  SELECT * FROM pg_stat_database WHERE datname = 'ticketing_system';
  ```

- [ ] Review customer support tickets
  - [ ] Inventory-related issues: _______
  - [ ] Checkout failures: _______
  - [ ] Unusual patterns: _______

### First Week
- [ ] Performance comparison
  | Metric | Before | After | Change |
  |--------|--------|-------|--------|
  | Avg order time | ___ms | ___ms | ___% |
  | Error rate | ___% | ___% | ___% |
  | Lock waits | ___ms | ___ms | ___ms |
  | Conflicts | N/A | ___% | N/A |

- [ ] Document any issues encountered
- [ ] Update runbooks if needed
- [ ] Share results with team

## Success Criteria

### Must Have ✅
- [x] Zero overselling incidents
- [ ] < 1% version conflict rate
- [ ] < 0.5% error rate increase
- [ ] < 50ms average lock wait time
- [ ] No customer complaints about inventory

### Nice to Have 🎯
- [ ] < 0.1% version conflict rate
- [ ] < 10ms average lock wait time
- [ ] Zero deadlocks
- [ ] Improved user confidence

## Rollback Procedure

If critical issues occur:

### Step 1: Revert Code (5 minutes)
```bash
# Stop application
sudo systemctl stop ticketing-api

# Revert to previous version
sudo cp /opt/ticketing/api-server.backup /opt/ticketing/api-server

# Start application
sudo systemctl start ticketing-api

# Verify
curl http://localhost:8080/health
```

### Step 2: Revert Database (Optional, 2 minutes)
```sql
-- Only if version column causes issues
BEGIN;
ALTER TABLE ticket_classes DROP COLUMN IF EXISTS version;
COMMIT;
```

### Step 3: Verify Rollback
- [ ] Application running
- [ ] Orders working
- [ ] No errors in logs
- [ ] Notify team of rollback

## Communication Plan

### Before Deployment
- [ ] Notify team in Slack: "Deploying concurrency fixes in 30 minutes"
- [ ] Create deployment ticket: DEPLOY-XXX
- [ ] Schedule deployment window: ___________

### During Deployment
- [ ] Update status: "Deployment in progress"
- [ ] Real-time updates on any issues
- [ ] ETA if delays occur

### After Deployment
- [ ] Success notification: "Deployment complete, monitoring..."
- [ ] Share initial metrics after 1 hour
- [ ] Document lessons learned

## Contacts

- **Primary Engineer**: _______
- **On-Call Engineer**: _______
- **Database Admin**: _______
- **Emergency Hotline**: _______

## Documentation

- Detailed docs: `CONCURRENCY_FIXES.md`
- Quick reference: `CONCURRENCY_QUICKREF.md`
- Visual guide: `CONCURRENCY_VISUAL_GUIDE.txt`
- Summary: `CONCURRENCY_RESOLUTION_SUMMARY.md`
- Tests: `internal/orders/concurrency_test.go`

## Notes

### Issues Encountered
```
Date: ___________
Issue: ___________
Resolution: ___________
```

### Lessons Learned
```
1. ___________
2. ___________
3. ___________
```

### Follow-up Actions
```
- [ ] ___________
- [ ] ___________
- [ ] ___________
```

---

**Deployment Date**: ___________
**Deployed By**: ___________
**Status**: [ ] Success [ ] Partial [ ] Rolled Back
**Sign-off**: ___________
