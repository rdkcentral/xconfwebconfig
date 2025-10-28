# Strategic Analysis: Best Files to Reach 60% Coverage

**Date:** October 28, 2025  
**Current Coverage:** 38.2% total  
**Target:** 60% minimum  
**Gap:** +21.8% needed  

---

## Coverage Analysis by File Size and Potential

### Main dataapi Package Analysis

| File | Lines | Current Coverage | Testable Functions | Estimated Gain | Priority |
|------|-------|------------------|-------------------|----------------|----------|
| **estb_firmware_handler.go** | 330 | **~75-93%** | Edge cases | **+2-3%** | 🔥 HIGH |
| **dataapi_common.go** | 301 | Mixed (0-100%) | 5 functions 0-38% | **+2-3%** | 🔥 HIGH |
| **feature_control_handler.go** | 389 | Partial | Handler improvements | **+1-2%** | 🟡 MEDIUM |
| **data_service_info.go** | 63 | **0%** | 3 simple handlers | **+0.5-1%** | ✅ QUICK WIN |

### estbfirmware Package Analysis (Currently 8.8%)

| File | Lines | Current Coverage | Testable Functions | Estimated Gain | Priority |
|------|-------|------------------|-------------------|----------------|----------|
| **estb_evaluation.go** | ~200 | **0%** | Evaluation logic | **+1-2%** | 🔥 HIGH |
| **estb_firmware_rule_eval.go** | ~500 | **41 functions** | Rule evaluation | **+2-3%** | 🔥 HIGH |
| **converter.go** | ~100 | **0%** | Conversion logic | **+1%** | 🟡 MEDIUM |
| **ip_mac_filter...go** (already done) | 445 | 8.8% | ✅ Done | - | ✅ COMPLETE |

### dcm/telemetry Package (Currently 28.1%)

| File | Lines | Current Coverage | Testable Functions | Estimated Gain | Priority |
|------|-------|------------------|-------------------|----------------|----------|
| **telemetry_profile.go** | 309 | **28.1%** | 15 functions at 0% | **+3-5%** | 🔥 HIGH |

### dcm/settings Package (Currently 18.2%)

| File | Lines | Current Coverage | Testable Functions | Estimated Gain | Priority |
|------|-------|------------------|-------------------|----------------|----------|
| **settings_profile.go** | ~100 | **18.2%** | 3 functions at 0% | **+1-2%** | 🟡 MEDIUM |

---

## TOP 5 HIGHEST IMPACT FILES TO REACH 60%

### 🥇 RANK 1: estb_firmware_handler.go (+2-3%)
**File:** `dataapi/estb_firmware_handler.go` (330 lines)

**Current State:**
- Most functions already 70-93% covered!
- Small improvements can yield big gains

**Functions to Test (edge cases):**
- `GetEstbFirmwareSwuBseHandler` - 93.5% → 100% (test edge cases)
- `GetFirmwareResponse` - 76.6% → 90%+ (test error paths)
- `GetEstbFirmwareSwuHandler` - 75.0% → 90%+ (test variations)
- `GetCheckMinFirmwareHandler` - 71.9% → 85%+ (test error conditions)

**Why This First:**
- ✅ Already mostly covered - just need edge cases
- ✅ Concrete, testable logic
- ✅ Can reach 90%+ file coverage
- ✅ **Expected: +2-3% total coverage**
- ✅ Lower effort than starting from 0%

**Estimated Effort:** 1-2 hours for 15-20 edge case tests

---

### 🥈 RANK 2: telemetry_profile.go (+3-5%)
**File:** `dataapi/dcm/telemetry/telemetry_profile.go` (309 lines)

**Current State:** 28.1% coverage

**Functions at 0% (High Value):**
- `GetTelemetryRuleForContext` - Rule matching logic
- `ProcessEntityRules` - Entity rule processing  
- `GetPermanentProfileByTelemetryRule` - Profile retrieval
- `ProcessTelemetryTwoRules` - Telemetry 2 rule processing
- `ExpireTemporaryTelemetryRules` - Rule expiration logic

**Why This Second:**
- ✅ Large file with many 0% functions
- ✅ Similar patterns to already-tested files
- ✅ Pure logic functions (testable)
- ✅ **Expected: +3-5% total coverage**

**Estimated Effort:** 3-4 hours for 25-30 tests

---

### 🥉 RANK 3: dataapi_common.go (+2-3%)
**File:** `dataapi/dataapi_common.go` (301 lines)

**Current State:** Mixed coverage

**Functions to Improve:**
- `AddContextFromTaggingService` - 38.5% → 80%+ (test more paths)
- `GetPartnerFromAccountServiceByHostMac` - 0% → Skip (requires HTTP)
- `AddGroupServiceContext` - 10% → 60%+ (test more scenarios)
- `AddGroupServiceFTContext` - 23.7% → 70%+ (test feature tags)
- `AddGroupServiceFeatureTags` - 0% → 60%+ (test tag addition logic)

**Why This Third:**
- ✅ Large file in main package
- ✅ Some functions partially covered - can improve
- ✅ **Expected: +2-3% total coverage**

**Estimated Effort:** 2-3 hours for 20-25 tests

---

### 🏅 RANK 4: estb_evaluation.go (+1-2%)
**File:** `dataapi/estbfirmware/estb_evaluation.go` (~200 lines)

**Current State:** 0% coverage

**Functions at 0%:**
- `NewEvaluationResult` - Constructor
- `AddAppliedFilters` - Filter tracking
- `DownloadLocationRoundRobinFilterFilter` - Load balancing logic

**Why This Fourth:**
- ✅ Pure logic functions
- ✅ Smaller file, manageable scope
- ✅ **Expected: +1-2% total coverage**

**Estimated Effort:** 1-2 hours for 10-15 tests

---

### 🏅 RANK 5: data_service_info.go (+0.5-1%)
**File:** `dataapi/data_service_info.go` (63 lines)

**Current State:** 0% coverage

**Functions at 0%:**
- `GetInfoRefreshAllHandler` - Simple handler
- `GetInfoRefreshHandler` - Simple handler
- `GetInfoStatistics` - Statistics handler

**Why This Fifth:**
- ✅ **QUICK WIN** - Small file
- ✅ Simple handlers (similar to already-tested)
- ✅ **Expected: +0.5-1% total coverage**

**Estimated Effort:** 30 minutes for 5-8 tests

---

## RECOMMENDED EXECUTION PLAN TO REACH 60%

### Phase 1: Quick Wins (Expected: +3-4%)
**Goal:** 38.2% → 41-42%

1. ✅ **estb_firmware_handler.go** (+2-3%) - 15-20 edge case tests
2. ✅ **data_service_info.go** (+0.5-1%) - 5-8 simple handler tests

**Time:** 2-3 hours  
**New Coverage:** ~41-42%

---

### Phase 2: Medium Impact (Expected: +5-6%)
**Goal:** 41-42% → 46-48%

3. ✅ **telemetry_profile.go** (+3-5%) - 25-30 tests for 0% functions
4. ✅ **estb_evaluation.go** (+1-2%) - 10-15 evaluation logic tests

**Time:** 4-6 hours  
**New Coverage:** ~46-48%

---

### Phase 3: Final Push (Expected: +4-5%)
**Goal:** 46-48% → 52-53%

5. ✅ **dataapi_common.go** (+2-3%) - 20-25 tests to improve partial coverage
6. ✅ **settings_profile.go** (+1-2%) - 10-12 tests for 0% functions

**Time:** 3-4 hours  
**New Coverage:** ~52-53%

---

### Phase 4: Additional Coverage (Expected: +7-10%)
**Goal:** 52-53% → 60%+

7. ✅ **estb_firmware_rule_eval.go** - More rule evaluation tests (+2-3%)
8. ✅ **dcm/logupload improvements** - Eval function (+2-3%)
9. ✅ **Edge cases in existing files** - Improve partial coverage (+2-3%)
10. ✅ **Additional helper functions** - Various files (+1-2%)

**Time:** 6-8 hours  
**Final Coverage:** ~60-62% ✅

---

## TOTAL EFFORT ESTIMATE

| Phase | Files | Tests | Time | Coverage Gain | Cumulative |
|-------|-------|-------|------|---------------|------------|
| Phase 1 | 2 files | 20-28 tests | 2-3 hours | +3-4% | **41-42%** |
| Phase 2 | 2 files | 35-45 tests | 4-6 hours | +5-6% | **46-48%** |
| Phase 3 | 2 files | 30-37 tests | 3-4 hours | +4-5% | **52-53%** |
| Phase 4 | 4+ files | 40-50 tests | 6-8 hours | +7-10% | **60-62%** ✅ |
| **TOTAL** | **10+ files** | **125-160 tests** | **15-21 hours** | **+19-25%** | **60%+** ✅ |

---

## WHY THIS ORDER?

### 1. **estb_firmware_handler.go FIRST** (Lowest Hanging Fruit)
- Already 70-93% covered
- Just need edge case tests
- Quick win with high impact
- Builds confidence and momentum

### 2. **telemetry_profile.go SECOND** (Biggest Single Opportunity)
- Large file with many 0% functions
- Similar patterns to already-tested code
- High value-to-effort ratio

### 3. **dataapi_common.go THIRD** (Improve Partial Coverage)
- Important main package file
- Functions already partially covered
- Can push many functions from 10-40% to 60-80%

### 4. **Small Files & Edge Cases** (Final Push)
- data_service_info.go - Quick win
- estb_evaluation.go - Manageable scope
- Various edge cases and improvements

---

## IMMEDIATE RECOMMENDATION

### 🔥 START WITH: estb_firmware_handler.go

**Why:**
1. ✅ Already mostly covered (70-93%) - need edge cases only
2. ✅ Immediate +2-3% impact
3. ✅ Quick win (1-2 hours)
4. ✅ Concrete testable logic
5. ✅ Builds momentum for bigger files

**Next:**
After estb_firmware_handler.go, do data_service_info.go for another quick win (+0.5-1%), then tackle telemetry_profile.go for the big +3-5% boost.

---

## COMPARISON TO ESTB_FIRMWARE_CONTEXT.GO

**estb_firmware_context.go status:**
- Already tested with good coverage on most functions
- Remaining functions are complex (require context setup)
- Would give +0-1% at most (diminishing returns)

**Better strategy:** Focus on files with 0% or low coverage first!

---

## SUMMARY

**To reach 60% coverage, test in this order:**

1. 🔥 **estb_firmware_handler.go** → +2-3% (START HERE!)
2. ✅ **data_service_info.go** → +0.5-1% (Quick win)
3. 🔥 **telemetry_profile.go** → +3-5% (Big impact)
4. 🟡 **estb_evaluation.go** → +1-2%
5. 🟡 **dataapi_common.go** → +2-3%
6. 🟡 **settings_profile.go** → +1-2%
7. 🟡 **Additional files** → +7-10%

**Expected Final Coverage: 60-62%** ✅

---

**Ready to start with estb_firmware_handler.go?** This will give us the quickest +2-3% boost!
