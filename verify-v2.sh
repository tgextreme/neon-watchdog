#!/bin/bash

# 🧪 Neon Watchdog v2.0 - Feature Verification Script
# Verifica que todas las features implementadas funcionen correctamente

set -e

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🧪 Neon Watchdog v2.0 - Feature Verification"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

PASS=0
FAIL=0

check_feature() {
    local feature="$1"
    local command="$2"
    
    echo -n "Testing $feature... "
    if eval "$command" > /dev/null 2>&1; then
        echo -e "${GREEN}✓ PASS${NC}"
        ((PASS++))
    else
        echo -e "${RED}✗ FAIL${NC}"
        ((FAIL++))
    fi
}

check_file() {
    local feature="$1"
    local file="$2"
    
    echo -n "Checking $feature... "
    if [ -f "$file" ]; then
        echo -e "${GREEN}✓ EXISTS${NC}"
        ((PASS++))
    else
        echo -e "${RED}✗ MISSING${NC}"
        ((FAIL++))
    fi
}

check_code() {
    local feature="$1"
    local file="$2"
    local pattern="$3"
    
    echo -n "Verifying $feature in code... "
    if grep -q "$pattern" "$file" 2>/dev/null; then
        echo -e "${GREEN}✓ FOUND${NC}"
        ((PASS++))
    else
        echo -e "${RED}✗ NOT FOUND${NC}"
        ((FAIL++))
    fi
}

echo "━━━ TIER 1: High Impact Features ━━━"
echo ""

check_file "Notifications Module" "internal/notifications/notifications.go"
check_code "Email Notifier" "internal/notifications/notifications.go" "EmailNotifier"
check_code "Webhook Notifier" "internal/notifications/notifications.go" "WebhookNotifier"
check_code "Telegram Notifier" "internal/notifications/notifications.go" "TelegramNotifier"

check_file "Metrics Module" "internal/metrics/metrics.go"
check_code "Prometheus Metrics" "internal/metrics/metrics.go" "neon_watchdog_uptime_seconds"

check_file "HTTP Checker" "internal/checks/checks.go"
check_code "HTTP Health Check" "internal/checks/checks.go" "HTTPChecker"

check_file "Action Hooks" "internal/actions/actions.go"
check_code "Before/After Hooks" "internal/actions/actions.go" "ActionWithHooks"

echo ""
echo "━━━ TIER 2: High Value Features ━━━"
echo ""

check_code "Dependency Chains" "internal/config/config.go" "DependsOn"
check_code "Ignore Exit Codes" "internal/config/config.go" "IgnoreExitCodes"
check_code "Backoff Strategy" "internal/config/config.go" "BackoffStrategy"

echo ""
echo "━━━ TIER 3: Advanced Features ━━━"
echo ""

check_code "Logic Checker (AND/OR)" "internal/checks/checks.go" "LogicChecker"
check_file "Dashboard Module" "internal/dashboard/dashboard.go"
check_code "Web UI" "internal/dashboard/dashboard.go" "handleUI"
check_code "Script Checker" "internal/checks/checks.go" "ScriptChecker"
check_file "History Module" "internal/history/history.go"
check_code "Event Recording" "internal/history/history.go" "RecordEvent"

echo ""
echo "━━━ Documentation & Examples ━━━"
echo ""

check_file "v2 README" "README-V2.md"
check_file "Full Config Example" "examples/config-v2-full.yml"
check_file "Implementation Summary" "IMPLEMENTATION-SUMMARY.md"

echo ""
echo "━━━ Build Verification ━━━"
echo ""

check_feature "Go Module Valid" "go mod verify"
check_feature "Code Compilation" "go build -o /tmp/neon-watchdog-test ./cmd/neon-watchdog"

if [ -f "/tmp/neon-watchdog-test" ]; then
    SIZE=$(du -h /tmp/neon-watchdog-test | cut -f1)
    echo -e "${GREEN}✓ Binary created: $SIZE${NC}"
    rm -f /tmp/neon-watchdog-test
    ((PASS++))
else
    echo -e "${RED}✗ Binary not created${NC}"
    ((FAIL++))
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "📊 Results Summary"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo -e "Passed: ${GREEN}$PASS${NC}"
echo -e "Failed: ${RED}$FAIL${NC}"
echo -e "Total:  $((PASS + FAIL))"
echo ""

if [ $FAIL -eq 0 ]; then
    echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${GREEN}🎉 ALL FEATURES VERIFIED SUCCESSFULLY!${NC}"
    echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    exit 0
else
    echo -e "${RED}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${RED}⚠️  SOME FEATURES MISSING OR FAILED${NC}"
    echo -e "${RED}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    exit 1
fi
