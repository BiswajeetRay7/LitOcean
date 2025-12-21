#!/usr/bin/env bash

# ============================================================
# 🌊⭐ LITOCEAN ⭐🌊
# Ocean + Star Subdomain Enumeration Framework
# Developed by Biswajeet Ray
# ============================================================

set -euo pipefail

# ---------------- COLORS ----------------
BLUE="\033[38;5;39m"
CYAN="\033[38;5;51m"
GREEN="\033[38;5;82m"
YELLOW="\033[38;5;220m"
PINK="\033[38;5;213m"
GRAY="\033[38;5;245m"
RESET="\033[0m"

# ---------------- MAIN ASCII ----------------
MAIN_ASCII="
${CYAN}
██╗     ██╗████████╗ ██████╗  ██████╗███████╗ █████╗ ███╗   ██╗
██║     ██║╚══██╔══╝██╔═══██╗██╔════╝██╔════╝██╔══██╗████╗  ██║
██║     ██║   ██║   ██║   ██║██║     █████╗  ███████║██╔██╗ ██║
██║     ██║   ██║   ██║   ██║██║     ██╔══╝  ██╔══██║██║╚██╗██║
███████╗██║   ██║   ╚██████╔╝╚██████╗███████╗██║  ██║██║ ╚████║
╚══════╝╚═╝   ╚═╝    ╚═════╝  ╚═════╝╚══════╝╚═╝  ╚═╝╚═╝  ╚═══╝
${RESET}
"

# ---------------- CUTE OCEAN ASCII ----------------
CUTE_ASCII="
${CYAN}
      ⭐      🌊        ⭐
   🌊      ⭐    🌊
        __/\\__
   ⭐   \\    /   🌊
        /____\\
   🌊   \\    /   ⭐
        /____\\
      ⭐      🌊        ⭐
${RESET}

${BLUE}🌊 L I T O C E A N 🌊${RESET}
${PINK}Cute Ocean Subdomain Hunter${RESET}
${GRAY}Developed by Biswajeet Ray${RESET}
"

# ---------------- ANIMATIONS ----------------
wave_loader() {
  local pid=$!
  local frames=("🌊   " " 🌊  " "  🌊 " "   🌊" "  🌊 " " 🌊  ")
  local i=0
  while kill -0 "$pid" 2>/dev/null; do
    echo -ne "${CYAN}[${frames[i]}] scanning ocean...${RESET}\r"
    i=$(( (i + 1) % ${#frames[@]} ))
    sleep 0.2
  done
  echo -ne "\r"
}

star_bar() {
  local total=$1
  local done=$2
  local stars=$(( done * 10 / total ))
  printf "${YELLOW}"
  for ((i=0;i<stars;i++)); do printf "⭐"; done
  for ((i=stars;i<10;i++)); do printf "·"; done
  printf "${RESET}"
}

# ---------------- HELP ----------------
usage() {
  clear
  echo -e "$MAIN_ASCII"
  echo -e "$CUTE_ASCII"
  echo
  echo -e "${BLUE}Usage:${RESET}"
  echo "  ./litocean.sh -d example.com"
  echo "  ./litocean.sh -l domains.txt"
  exit 0
}

# ---------------- DEPENDENCIES ----------------
need() { command -v "$1" >/dev/null 2>&1; }

install_go_tool() {
  echo -e "${YELLOW}✨ Installing $1...${RESET}"
  go install "$2@latest"
}

ensure_tools() {
  need go || { echo "Go missing"; exit 1; }
  for t in subfinder amass assetfinder chaos findomain httpx anew jq curl; do
    if ! need "$t"; then
      case "$t" in
        subfinder) install_go_tool subfinder github.com/projectdiscovery/subfinder/v2/cmd/subfinder ;;
        amass) install_go_tool amass github.com/owasp-amass/amass/v4/cmd/amass ;;
        assetfinder) install_go_tool assetfinder github.com/tomnomnom/assetfinder ;;
        chaos) install_go_tool chaos github.com/projectdiscovery/chaos-client/cmd/chaos ;;
        httpx) install_go_tool httpx github.com/projectdiscovery/httpx/cmd/httpx ;;
        anew) install_go_tool anew github.com/tomnomnom/anew ;;
        findomain)
          echo -e "${YELLOW}✨ Installing findomain...${RESET}"
          curl -sL https://github.com/findomain/findomain/releases/latest/download/findomain-linux -o findomain
          chmod +x findomain
          sudo mv findomain /usr/local/bin/
          ;;
        *) echo "Missing $t"; exit 1 ;;
      esac
    fi
  done
}

# ---------------- ARGS ----------------
DOMAIN=""
LIST=""
while getopts "d:l:h" opt; do
  case $opt in
    d) DOMAIN="$OPTARG" ;;
    l) LIST="$OPTARG" ;;
    h) usage ;;
  esac
done
[[ -z "$DOMAIN" && -z "$LIST" ]] && usage

# ---------------- START ----------------
clear
echo -e "$MAIN_ASCII"
echo -e "$CUTE_ASCII"

ensure_tools

WORKDIR="litocean-output"
SUBS="$WORKDIR/subs.txt"
ALIVE="$WORKDIR/alive.txt"
mkdir -p "$WORKDIR"
> "$SUBS"
> "$ALIVE"

TOOLS_TOTAL=7
TOOLS_DONE=0

run_step() {
  local name="$1"
  local cmd="$2"
  echo -ne "${BLUE}➜ $name ${GRAY}"
  eval "$cmd" & wave_loader
  TOOLS_DONE=$((TOOLS_DONE + 1))
  echo -ne "\r${GREEN}✔ $name ${RESET} "
  star_bar "$TOOLS_TOTAL" "$TOOLS_DONE"
  echo
}

enumerate() {
  local d="$1"
  echo -e "\n${PINK}🌊 Exploring ocean for: $d 🌊${RESET}"

  run_step "crt.sh" \
    "curl -s 'https://crt.sh/?q=%25.$d&output=json' | jq -r '.[].name_value' | sed 's/\*\.//g' | anew '$SUBS'"

  run_step "wayback" \
    "curl -s 'https://web.archive.org/cdx/search/cdx?url=*.$d/*&output=json&fl=original&collapse=urlkey' | jq -r '.[][].?' | sed -E 's_https?://([^/]+)/.*_\\1_' | grep '$d' | anew '$SUBS'"

  run_step "subfinder" "subfinder -d '$d' -all -silent | anew '$SUBS'"
  run_step "amass" "amass enum -passive -d '$d' | anew '$SUBS'"
  run_step "assetfinder" "assetfinder --subs-only '$d' | anew '$SUBS'"
  run_step "chaos" "chaos -d '$d' -silent | anew '$SUBS'"
  run_step "findomain" "findomain -t '$d' -q | anew '$SUBS'"
}

if [[ -n "$DOMAIN" ]]; then
  enumerate "$DOMAIN"
else
  while read -r d; do enumerate "$d"; done < "$LIST"
fi

echo -e "\n${CYAN}🌟 Total subdomains: $(wc -l < "$SUBS")${RESET}"
echo -e "${BLUE}🌊 Checking alive hosts...${RESET}"

cat "$SUBS" | httpx -silent -threads 200 | anew "$ALIVE"

echo -e "${GREEN}⭐ Alive hosts: $(wc -l < "$ALIVE")${RESET}"
echo -e "${PINK}✨ Results saved in $WORKDIR ✨${RESET}"
