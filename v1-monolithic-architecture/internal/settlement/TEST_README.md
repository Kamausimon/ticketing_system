# Settlement Routes Test Suite

This test suite provides comprehensive testing for all settlement-related routes in the ticketing system.

## Overview

The settlement routes handle the calculation, creation, approval, and processing of settlements (payouts to event organizers). Since these operations involve complex financial calculations and state transitions, testing via Postman directly can be challenging. These automated tests provide a reliable way to verify all settlement functionality.

## Running the Tests

To run all settlement tests:

```bash
cd /home/kamau/projects/ticketing_system
go test ./internal/settlement/... -v
```

To run a specific test:

```bash
go test ./internal/settlement/... -v -run TestCalculateEventSettlement
```

## Test Coverage

### Calculation & Preview Routes

1. **TestCalculateEventSettlement** - Tests calculating settlement amounts for a completed event
   - Verifies gross amount, platform fees, gateway fees, and net amount calculations
   - Ensures proper fee deductions

2. **TestCalculateEventSettlement_NotCompleted** - Tests that settlements can't be calculated for non-completed events
   - Validates business rules around event completion

3. **TestGetSettlementPreview** - Tests settlement preview functionality
   - Allows organizers to see estimated settlement amounts

4. **TestValidateSettlementEligibility** - Tests eligibility validation for settlements
   - Checks if an event meets all requirements for settlement

### Settlement Batch Operations

5. **TestCreateSettlementBatch** - Tests creating a new settlement batch
   - Verifies settlement record creation
   - Validates settlement items for organizers
   - Ensures payment records are properly linked

6. **TestGetSettlement** - Tests retrieving a specific settlement by ID
   - Validates settlement data retrieval

7. **TestListSettlements** - Tests listing settlements with pagination
   - Verifies pagination functionality
   - Checks response structure

8. **TestListSettlementsWithFilters** - Tests filtering settlements by status
   - Validates query parameter filtering

### Settlement State Management

9. **TestApproveSettlement** - Tests approving a pending settlement
   - Verifies status changes from pending to ready_to_process
   - Validates approval tracking (who approved and when)
   - Tests holding period requirements

10. **TestCancelSettlement** - Tests canceling a pending settlement
    - Verifies cancellation workflow
    - Validates reason tracking

11. **TestWithholdSettlement** - Tests withholding a settlement
    - Used when there are disputes or compliance issues
    - Validates withholding reason tracking

12. **TestGetPendingSettlements** - Tests retrieving all pending settlements
    - Useful for admin dashboards

### Error Handling

13. **TestInvalidEventID** - Tests handling of invalid event IDs
    - Validates input validation and error responses

14. **TestInvalidSettlementID** - Tests handling of invalid settlement IDs
    - Ensures proper error handling for malformed IDs

15. **TestNonExistentSettlement** - Tests attempting to retrieve non-existent settlements
    - Validates 404 responses

16. **TestCreateSettlementBatch_InvalidRequest** - Tests invalid request body handling
    - Validates JSON parsing and error responses

## Test Data Setup

Each test automatically sets up:
- A test account
- A user
- An organizer with verified payout account
- A completed event with ticket sales
- Payment records (customer payments, platform fees, gateway fees)

This ensures each test runs independently with isolated data.

## Key Endpoints Tested

| Method | Route | Description |
|--------|-------|-------------|
| GET | `/settlements/calculate/event/{id}` | Calculate settlement for specific event |
| GET | `/settlements/preview` | Get settlement preview |
| GET | `/settlements/eligibility/event/{id}` | Check settlement eligibility |
| POST | `/settlements/batch` | Create new settlement batch |
| GET | `/settlements/{id}` | Get specific settlement |
| GET | `/settlements` | List settlements with filters |
| POST | `/settlements/{id}/approve` | Approve settlement |
| POST | `/settlements/{id}/cancel` | Cancel settlement |
| POST | `/settlements/{id}/withhold` | Withhold settlement |
| GET | `/settlements/pending` | Get pending settlements |

## Settlement Flow Tested

1. **Event Completion** → Event must be completed
2. **Calculation** → Calculate gross amount, fees, deductions
3. **Batch Creation** → Create settlement batch with items
4. **Holding Period** → Wait for holding period (default 7 days)
5. **Approval** → Admin approves settlement
6. **Processing** → Settlement is processed for payout
7. **Completion** → Funds transferred to organizers

## Notes

- All tests use an in-memory SQLite database for isolation
- Tests automatically migrate required tables
- Each test is independent and doesn't affect others
- The `stringPtr` helper function is used to create string pointers for optional fields
- Holding periods are bypassed in tests by manually adjusting the `earliest_payout_date`

## Common Test Patterns

### Creating a Settlement
```go
reqBody := CreateSettlementBatchRequest{
    Description:       "Test Settlement",
    Frequency:         models.SettlementPostEvent,
    Trigger:           models.TriggerEventCompletion,
    PeriodStartDate:   time.Now().Add(-30 * 24 * time.Hour),
    PeriodEndDate:     time.Now(),
    HoldingPeriodDays: 7,
    InitiatedByUserID: userID,
    EventID:           &eventID,
}
settlement, err := service.CreateSettlementBatch(reqBody)
```

### Testing HTTP Endpoints
```go
req, err := http.NewRequest("GET", "/settlements/123", nil)
rr := httptest.NewRecorder()
router := mux.NewRouter()
router.HandleFunc("/settlements/{id}", handler.GetSettlement)
router.ServeHTTP(rr, req)

assert.Equal(t, http.StatusOK, rr.Code)
```

## Troubleshooting

If tests fail:

1. **Unique Constraint Errors** - The batch ID is generated using Unix timestamp. If multiple settlements are created in the same second, they'll have duplicate IDs. Tests are designed to avoid this.

2. **Holding Period Errors** - Tests bypass holding periods by manually updating the `earliest_payout_date` field after creating settlements.

3. **Missing Tables** - Ensure all models are included in the AutoMigrate call in `setupTestDB()`

## Future Enhancements

Potential additional tests:
- Settlement processing with payment gateway integration (currently tested at service level)
- Retry logic for failed settlements
- Settlement reports generation
- Settlement history tracking
- Bulk settlement operations
