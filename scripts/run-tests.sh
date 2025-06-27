#!/bin/bash

# Test runner script for Internal Transfers Application
# This script provides various testing options and configurations

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default configuration
TEST_TYPE="all"
COVERAGE=false
VERBOSE=false
RACE=false
INTEGRATION=false
CLEANUP=true

# Helper functions
print_usage() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  -t, --type TYPE       Test type: unit, integration, all (default: all)"
    echo "  -c, --coverage        Generate coverage report"
    echo "  -v, --verbose         Verbose output"
    echo "  -r, --race            Enable race detection"
    echo "  -i, --integration     Run integration tests (requires database)"
    echo "  --no-cleanup          Skip cleanup after tests"
    echo "  -h, --help            Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 -t unit -c         # Run unit tests with coverage"
    echo "  $0 -t integration -v  # Run integration tests with verbose output"
    echo "  $0 -c -r              # Run all tests with coverage and race detection"
}

print_banner() {
    echo -e "${BLUE}"
    echo "=================================================="
    echo "  Internal Transfers Application - Test Runner"
    echo "=================================================="
    echo -e "${NC}"
}

print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -t|--type)
            TEST_TYPE="$2"
            shift 2
            ;;
        -c|--coverage)
            COVERAGE=true
            shift
            ;;
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        -r|--race)
            RACE=true
            shift
            ;;
        -i|--integration)
            INTEGRATION=true
            shift
            ;;
        --no-cleanup)
            CLEANUP=false
            shift
            ;;
        -h|--help)
            print_usage
            exit 0
            ;;
        *)
            print_error "Unknown option $1"
            print_usage
            exit 1
            ;;
    esac
done

# Validate test type
if [[ "$TEST_TYPE" != "unit" && "$TEST_TYPE" != "integration" && "$TEST_TYPE" != "all" ]]; then
    print_error "Invalid test type: $TEST_TYPE"
    print_usage
    exit 1
fi

print_banner

# Set up environment
setup_environment() {
    print_info "Setting up test environment..."
    
    # Create necessary directories
    mkdir -p coverage reports
    
    # Set environment variables
    export GO_ENV=test
    export TEST_ENV=$TEST_TYPE
    
    if [[ "$INTEGRATION" == "true" || "$TEST_TYPE" == "integration" || "$TEST_TYPE" == "all" ]]; then
        export RUN_INTEGRATION_TESTS=true
        export TEST_DATABASE_URL=${TEST_DATABASE_URL:-"postgres://user:password@localhost:5432/internal_transfers_test?sslmode=disable"}
        print_info "Integration tests enabled"
        print_info "Database URL: $TEST_DATABASE_URL"
    fi
}

# Check prerequisites
check_prerequisites() {
    print_info "Checking prerequisites..."
    
    # Check Go installation
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed or not in PATH"
        exit 1
    fi
    
    # Check Go version
    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    print_info "Go version: $GO_VERSION"
    
    # Check if dependencies are installed
    if [[ ! -f "go.mod" ]]; then
        print_error "go.mod not found. Run this script from the project root."
        exit 1
    fi
    
    # Download dependencies
    print_info "Downloading dependencies..."
    go mod download
    
    # Check database connection for integration tests
    if [[ "$INTEGRATION" == "true" || "$TEST_TYPE" == "integration" || "$TEST_TYPE" == "all" ]]; then
        check_database_connection
    fi
}

# Check database connection
check_database_connection() {
    print_info "Checking database connection..."
    
    # Try to connect to the database
    if command -v psql &> /dev/null; then
        if psql "$TEST_DATABASE_URL" -c "SELECT 1;" &> /dev/null; then
            print_success "Database connection successful"
        else
            print_warning "Cannot connect to test database. Integration tests may fail."
            print_info "Ensure PostgreSQL is running and the test database exists."
        fi
    else
        print_warning "psql not found. Cannot verify database connection."
    fi
}

# Build test flags
build_test_flags() {
    FLAGS=""
    
    if [[ "$VERBOSE" == "true" ]]; then
        FLAGS="$FLAGS -v"
    fi
    
    if [[ "$RACE" == "true" ]]; then
        FLAGS="$FLAGS -race"
    fi
    
    if [[ "$COVERAGE" == "true" ]]; then
        FLAGS="$FLAGS -coverprofile=coverage/coverage.out"
    fi
    
    # Add timeout
    FLAGS="$FLAGS -timeout=5m"
    
    echo "$FLAGS"
}

# Run unit tests
run_unit_tests() {
    print_info "Running unit tests..."
    
    local flags=$(build_test_flags)
    
    if go test ./tests/unit/... $flags; then
        print_success "Unit tests passed"
        return 0
    else
        print_error "Unit tests failed"
        return 1
    fi
}

# Run integration tests
run_integration_tests() {
    print_info "Running integration tests..."
    
    local flags=$(build_test_flags)
    
    # Update coverage profile name for integration tests
    if [[ "$COVERAGE" == "true" ]]; then
        flags=$(echo "$flags" | sed 's/coverage.out/coverage-integration.out/')
    fi
    
    if go test ./tests/integration/... $flags; then
        print_success "Integration tests passed"
        return 0
    else
        print_error "Integration tests failed"
        return 1
    fi
}

# Run all tests
run_all_tests() {
    print_info "Running all tests..."
    
    local unit_result=0
    local integration_result=0
    
    # Run unit tests
    run_unit_tests || unit_result=$?
    
    # Run integration tests if enabled
    if [[ "$INTEGRATION" == "true" ]]; then
        run_integration_tests || integration_result=$?
    fi
    
    # Check results
    if [[ $unit_result -eq 0 && $integration_result -eq 0 ]]; then
        print_success "All tests passed"
        return 0
    else
        print_error "Some tests failed"
        return 1
    fi
}

# Generate coverage report
generate_coverage_report() {
    if [[ "$COVERAGE" != "true" ]]; then
        return 0
    fi
    
    print_info "Generating coverage report..."
    
    # Combine coverage files if both exist
    if [[ -f "coverage/coverage.out" && -f "coverage/coverage-integration.out" ]]; then
        print_info "Combining coverage reports..."
        echo "mode: atomic" > coverage/combined.out
        tail -n +2 coverage/coverage.out >> coverage/combined.out
        tail -n +2 coverage/coverage-integration.out >> coverage/combined.out
        
        # Generate HTML report from combined coverage
        go tool cover -html=coverage/combined.out -o coverage/coverage.html
        
        # Show coverage summary
        go tool cover -func=coverage/combined.out | grep total
    elif [[ -f "coverage/coverage.out" ]]; then
        # Generate HTML report
        go tool cover -html=coverage/coverage.out -o coverage/coverage.html
        
        # Show coverage summary
        go tool cover -func=coverage/coverage.out | grep total
    fi
    
    if [[ -f "coverage/coverage.html" ]]; then
        print_success "Coverage report generated: coverage/coverage.html"
    fi
}

# Cleanup function
cleanup() {
    if [[ "$CLEANUP" == "true" ]]; then
        print_info "Cleaning up..."
        
        # Clean up temporary files
        find . -name "*.test" -delete 2>/dev/null || true
        find . -name "*.prof" -delete 2>/dev/null || true
        
        print_info "Cleanup completed"
    fi
}

# Main execution
main() {
    setup_environment
    check_prerequisites
    
    local exit_code=0
    
    case $TEST_TYPE in
        unit)
            run_unit_tests || exit_code=$?
            ;;
        integration)
            export RUN_INTEGRATION_TESTS=true
            run_integration_tests || exit_code=$?
            ;;
        all)
            run_all_tests || exit_code=$?
            ;;
    esac
    
    generate_coverage_report
    cleanup
    
    # Final summary
    if [[ $exit_code -eq 0 ]]; then
        print_success "Test execution completed successfully!"
    else
        print_error "Test execution failed!"
    fi
    
    exit $exit_code
}

# Trap to ensure cleanup runs on exit
trap cleanup EXIT

# Run main function
main "$@"