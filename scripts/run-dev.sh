#!/bin/bash

# Comprehensive Development Script for stratocraft.dev
# Runs Tailwind CSS watch, Templ watch, and Air in parallel with proper cleanup

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# PID tracking for cleanup
TAILWIND_PID=""
TEMPL_PID=""
AIR_PID=""

# Function to display usage
usage() {
    echo "Usage: $0 [start|stop|restart|status]"
    echo ""
    echo "Commands:"
    echo "  start   - Start all development processes (default)"
    echo "  stop    - Stop all running development processes"
    echo "  restart - Stop and start all processes"
    echo "  status  - Show status of development processes"
    echo ""
    echo "Processes managed:"
    echo "  • Tailwind CSS watch (background)"
    echo "  • Templ template watch (background)"
    echo "  • Air Go hot reload (foreground)"
}

# Function to log messages
log() {
    echo -e "${GREEN}[$(date +'%H:%M:%S')]${NC} $*"
}

log_warn() {
    echo -e "${YELLOW}[$(date +'%H:%M:%S')] WARNING:${NC} $*"
}

log_error() {
    echo -e "${RED}[$(date +'%H:%M:%S')] ERROR:${NC} $*"
}

# Function to check if a command exists
check_command() {
    if ! command -v "$1" &> /dev/null; then
        log_error "$1 is not installed or not in PATH"
        return 1
    fi
    return 0
}

# Function to check if process is running
is_process_running() {
    local pid=$1
    if [ -n "$pid" ] && kill -0 "$pid" 2>/dev/null; then
        return 0
    fi
    return 1
}

# Function to stop a process gracefully
stop_process() {
    local pid=$1
    local name=$2
    
    if is_process_running "$pid"; then
        log "Stopping $name (PID: $pid)..."
        kill -TERM "$pid" 2>/dev/null || true
        
        # Wait up to 5 seconds for graceful shutdown
        local count=0
        while is_process_running "$pid" && [ $count -lt 5 ]; do
            sleep 1
            count=$((count + 1))
        done
        
        # Force kill if still running
        if is_process_running "$pid"; then
            log_warn "Force killing $name (PID: $pid)..."
            kill -KILL "$pid" 2>/dev/null || true
        fi
    fi
}

# Function to cleanup all processes
cleanup() {
    log "Cleaning up development processes..."
    
    stop_process "$TAILWIND_PID" "Tailwind CSS watch"
    stop_process "$TEMPL_PID" "Templ watch"
    stop_process "$AIR_PID" "Air"
    
    # Also clean up any processes by name (in case PIDs are lost)
    pkill -f "tailwindcss.*--watch" 2>/dev/null || true
    pkill -f "templ.*--watch" 2>/dev/null || true
    pkill -f "air" 2>/dev/null || true
    
    log "Cleanup complete"
    exit 0
}

# Function to check environment variables
check_environment() {
    if [ -z "$GITHUB_TOKEN" ]; then
        log_warn "GITHUB_TOKEN environment variable is not set."
        echo "         To avoid GitHub API rate limiting, please:"
        echo "         1. Create a Personal Access Token at: https://github.com/settings/tokens"
        echo "         2. Export it: export GITHUB_TOKEN=your_token_here"
        echo "         3. Run this script again"
        echo ""
        echo "         Continuing without authentication (rate limited to 60 requests/hour)..."
        echo ""
    fi

    if [ -z "$GITHUB_WEBHOOK_SECRET" ]; then
        log_warn "GITHUB_WEBHOOK_SECRET environment variable is not set."
        echo "         This is only needed for GitHub webhooks. To set up:"
        echo "         1. Generate a secret: openssl rand -hex 32"
        echo "         2. Export it: export GITHUB_WEBHOOK_SECRET=your_secret_here"
        echo "         3. Configure it in your GitHub repo webhook settings"
        echo ""
    fi
}

# Function to check prerequisites
check_prerequisites() {
    log "Checking prerequisites..."
    
    local missing=0
    
    # Check required commands
    for cmd in templ npx air; do
        if ! check_command "$cmd"; then
            missing=1
        fi
    done
    
    if [ $missing -eq 1 ]; then
        log_error "Missing required dependencies. Please install:"
        echo "  • templ: go install github.com/a-h/templ/cmd/templ@latest"
        echo "  • air: go install github.com/cosmtrek/air@latest"
        echo "  • npx: comes with Node.js - install from https://nodejs.org/"
        exit 1
    fi
    
    # Check if Tailwind CSS is available
    if ! npx tailwindcss --help &>/dev/null; then
        log_error "Tailwind CSS is not installed. Please run: npm install"
        exit 1
    fi
    
    # Check required files
    if [ ! -f "public/css/style.css" ]; then
        log_error "public/css/style.css not found"
        exit 1
    fi
    
    if [ ! -f ".air.toml" ]; then
        log_error ".air.toml configuration file not found"
        exit 1
    fi
    
    log "Prerequisites check passed ✓"
}

# Function to show process status
show_status() {
    echo -e "${BLUE}Development Process Status:${NC}"
    echo ""
    
    # Check for running processes
    local tailwind_running=$(pgrep -f "tailwindcss.*--watch" 2>/dev/null || echo "")
    local templ_running=$(pgrep -f "templ.*--watch" 2>/dev/null || echo "")
    local air_running=$(pgrep -f "air" 2>/dev/null || echo "")
    
    if [ -n "$tailwind_running" ]; then
        echo -e "  ${GREEN}✓${NC} Tailwind CSS watch (PID: $tailwind_running)"
    else
        echo -e "  ${RED}✗${NC} Tailwind CSS watch"
    fi
    
    if [ -n "$templ_running" ]; then
        echo -e "  ${GREEN}✓${NC} Templ watch (PID: $templ_running)"
    else
        echo -e "  ${RED}✗${NC} Templ watch"
    fi
    
    if [ -n "$air_running" ]; then
        echo -e "  ${GREEN}✓${NC} Air hot reload (PID: $air_running)"
    else
        echo -e "  ${RED}✗${NC} Air hot reload"
    fi
    
    echo ""
    echo "URLs:"
    echo "  • Application: http://localhost:8080"
    echo "  • Webhook endpoint: http://localhost:8080/webhook/github"
}

# Function to stop all development processes
stop_development() {
    log "Stopping all development processes..."
    
    # Kill processes by name
    pkill -f "tailwindcss.*--watch" 2>/dev/null && log "Stopped Tailwind CSS watch" || true
    pkill -f "templ.*--watch" 2>/dev/null && log "Stopped Templ watch" || true
    pkill -f "air" 2>/dev/null && log "Stopped Air" || true
    
    log "All development processes stopped"
}

# Function to start all development processes
start_development() {
    log "Starting stratocraft.dev development environment..."
    
    check_prerequisites
    check_environment
    
    # Initial build
    log "Running initial build..."
    templ generate
    npx tailwindcss -i ./public/css/style.css -o ./public/css/site.css --minify
    
    # Start Tailwind CSS watch in background
    log "Starting Tailwind CSS watch..."
    ./scripts/tailwind-watch.sh &
    TAILWIND_PID=$!
    sleep 1
    if is_process_running "$TAILWIND_PID"; then
        log "Tailwind CSS watch started (PID: $TAILWIND_PID)"
    else
        log_error "Failed to start Tailwind CSS watch"
        exit 1
    fi
    
    # Start Templ watch in background
    log "Starting Templ watch..."
    ./scripts/templ-watch.sh &
    TEMPL_PID=$!
    sleep 1
    if is_process_running "$TEMPL_PID"; then
        log "Templ watch started (PID: $TEMPL_PID)"
    else
        log_error "Failed to start Templ watch"
        cleanup
        exit 1
    fi
    
    # Give background processes time to initialize
    sleep 2
    
    # Show status
    echo ""
    echo -e "${GREEN}🚀 Development environment ready!${NC}"
    echo ""
    echo -e "${BLUE}Running processes:${NC}"
    echo "  • Tailwind CSS watch (background)"
    echo "  • Templ template watch (background)"
    echo "  • Air Go hot reload (starting...)"
    echo ""
    echo -e "${BLUE}URLs:${NC}"
    echo "  • Application: http://localhost:8080"
    echo "  • Webhook endpoint: http://localhost:8080/webhook/github"
    echo ""
    echo -e "${YELLOW}Press Ctrl+C to stop all processes${NC}"
    echo ""
    
    # Start Air in foreground (this will block until Ctrl+C)
    log "Starting Air hot reload..."
    air
}

# Set up signal handling for cleanup
trap cleanup SIGINT SIGTERM

# Main script logic
case "${1:-start}" in
    start)
        start_development
        ;;
    stop)
        stop_development
        ;;
    restart)
        stop_development
        sleep 2
        start_development
        ;;
    status)
        show_status
        ;;
    help|--help|-h)
        usage
        ;;
    *)
        log_error "Unknown command: $1"
        usage
        exit 1
        ;;
esac 