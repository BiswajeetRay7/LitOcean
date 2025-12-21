#!/usr/bin/env bash

# ============================================================
# 🌊⭐ LITOCEAN ⭐🌊
# Fast Parallel Subdomain Enumeration (Bash)
# Developed by Biswajeet Ray
# ============================================================

set -u

# ---------------- COLORS ----------------
CYAN="\033[36m"
GREEN="\033[32m"
BLUE="\033[34m"
YELLOW="\033[33m"
GRAY="\033[90m"
RED="\033[31m"
RESET="\033[0m"

# ---------------- TRAP (CTRL+C) ----------------
trap 'echo -e "\n${RED}[!] Interrupted by user. Killing processes...${RESET}"; kill $(jobs -p) 2>/dev/null; exit 1' SIGINT SIGTERM

# ---------------- ASCII ART ----------------
show_banner() {
    clear
    cat << "EOF"
██╗      ██╗████████╗ ██████╗  ██████╗███████╗ █████╗ ███╗   ██╗
██║      ██║╚══██╔══╝██╔═══██╗██╔════╝██╔════╝██╔══██╗████╗  ██║
██║      ██║   ██║   ██║   ██║██║      █████╗  ███████║██╔██╗ ██║
██║      ██║   ██║   ██║   ██║██║     ██╔══╝ ██╔══██║██║╚██╗██║
███████╗██║   ██║   ╚██████╔╝╚██████╗███████╗██║  ██║██║ ╚████║
╚══════╝╚═╝   ╚═╝    ╚═════╝  ╚═════╝╚══════╝╚═╝  ╚═╝╚═╝   ╚═══╝

🌊 Ocean + Star Subdomain Hunter v2.0
Developed by Biswajeet Ray
EOF
}

# ---------------- HELP ----------------
usage() {
  show_banner
  echo
  echo -e "${YELLOW}Usage:${RESET}"
  echo "  $0 -d example.com      (Single Target)"
  echo "  $0 -l domains.txt      (List of Targets)"
  echo
  exit 0
}

# ---------------- DEPENDENCY CHECK ----------------
need() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo -e "${RED}[!] Critical: Missing dependency '$1'. Please install it.${RESET}"
    exit 1
  fi
}

# Check all required tools
for dep in curl jq subfinder amass assetfinder findomain httpx anew; do
  need "$dep"
done

# ---------------- ARGS PARSING ----------------
DOMAIN=""
LIST=""

while getopts "d:l:h" opt; do
  case "$opt" in
    d) DOMAIN="$OPTARG" ;;
    l) LIST="$OPTARG" ;;
    h) usage ;;
    *) usage ;;
  esac
done

[[ -z "$DOMAIN" && -z "$LIST" ]] && usage

show_banner

# ---------------- HELPER FUNCTIONS ----------------

# Query CRT.SH with timeout and error handling
crtsh() {
  local target="$1"
  # Timeout set to 15 seconds to prevent hanging
  curl -s --max-time 20 "https://crt.sh/?q=%25.$target&output=json" \
    | jq -r '.. | .name_value? // empty' 2>/dev/null \
    | sed 's/\*\.//g' \
    | grep -i "$target" || true
}

# Query Wayback Machine with timeout
wayback() {
  local target="$1"
  curl -s --max-time 20 "https://web.archive.org/cdx/search/cdx?url=*.$target/*&output=json" \
    | jq -r '.[1:][].[]' 2>/dev/null \
    | sed -E 's_https?://([^/]+)/.*_\1_' \
    | grep -i "$target" || true
}

# ---------------- CORE LOGIC ----------------

process_target() {
    local target="$1"
    
    # Create unique workspace per domain
    local BASE_DIR="litocean_results/${target}"
    local TMP_DIR="$BASE_DIR/tmp"
    local SUBS_FILE="$BASE_DIR/all_subs.txt"
    local ALIVE_FILE="$BASE_DIR/alive.txt"

    mkdir -p "$TMP_DIR"
    : > "$SUBS_FILE"

    echo -e "\n${CYAN}==========================================${RESET}"
    echo -e "${CYAN}[*] Targeting Domain: ${YELLOW}$target${RESET}"
    echo -e "${CYAN}==========================================${RESET}"

    echo -e "${BLUE}[*] Launching parallel enumeration tools...${RESET}"

    # 1. CRT.SH
    { 
      crtsh "$target" | anew "$TMP_DIR/crt.txt" >/dev/null
    } &

    # 2. Wayback
    { 
      wayback "$target" | anew "$TMP_DIR/wayback.txt" >/dev/null 
    } &

    # 3. Subfinder
    { 
      subfinder -d "$target" -all -silent 2>/dev/null > "$TMP_DIR/subfinder.txt" 
    } &

    # 4. Assetfinder
    { 
      assetfinder --subs-only "$target" > "$TMP_DIR/assetfinder.txt" 
    } &

    # 5. Findomain
    { 
      findomain -t "$target" -q 2>/dev/null > "$TMP_DIR/findomain.txt" 
    } &

    # 6. Amass (With timeout of 5 minutes to prevent stalling)
    { 
      timeout 5m amass enum -passive -d "$target" -timeout 5 2>/dev/null > "$TMP_DIR/amass.txt" || true
    } &

    # Wait for all background jobs to finish
    wait

    echo -e "${BLUE}[*] Aggregating results...${RESET}"
    
    # Merge all results using anew to remove duplicates
    cat "$TMP_DIR"/*.txt 2>/dev/null | anew "$SUBS_FILE" >/dev/null

    # Check if we found anything
    local count=$(wc -l < "$SUBS_FILE")
    if [[ "$count" -eq 0 ]]; then
        echo -e "${RED}[!] No subdomains found for $target.${RESET}"
        rm -rf "$TMP_DIR"
        return
    fi
    echo -e "${GREEN}[+] Unique Subdomains found: $count${RESET}"

    # Probing with HTTPX
    echo -e "${BLUE}[*] Probing for alive hosts (HTTP/HTTPS)...${RESET}"
    httpx -l "$SUBS_FILE" -silent -threads 100 -timeout 5 | anew "$ALIVE_FILE" >/dev/null

    local alive_count=$(wc -l < "$ALIVE_FILE")
    
    # Summary
    echo -e "${GREEN}------------------------------------------${RESET}"
    echo -e "${GREEN}[✓] Scan Completed for $target${RESET}"
    echo -e "    - Total Subs : $count"
    echo -e "    - Alive URLs : $alive_count"
    echo -e "    - File Saved : $ALIVE_FILE"
    echo -e "${GREEN}------------------------------------------${RESET}"

    # Cleanup Tmp
    rm -rf "$TMP_DIR"
}

# ---------------- EXECUTION FLOW ----------------

if [[ -n "$DOMAIN" ]]; then
    process_target "$DOMAIN"
elif [[ -n "$LIST" ]]; then
    if [[ ! -f "$LIST" ]]; then
        echo -e "${RED}[!] File not found: $LIST${RESET}"
        exit 1
    fi
    echo -e "${BLUE}[*] Reading from list: $LIST${RESET}"
    while read -r line; do
        # cleanup whitespace
        target=$(echo "$line" | xargs)
        if [[ -n "$target" ]]; then
            process_target "$target"
        fi
    done < "$LIST"
fi

echo -e "\n${CYAN}🌊 LITOCEAN Execution Finished.${RESET}"
