#!/bin/bash

RESET="\033[0m"          # Normal Colour
RED="\033[0;31m"         # Error / Issues
GREEN="\033[0;32m"       # Successful
BOLD="\033[01;01m"       # Highlight
WHITE="\033[1;37m"       # Bold Text
YELLOW="\033[1;33m"      # Warnings and Info
BLINK="\033[5m"          # Blinking Effect
CYAN="\033[0;36m"        # Info
MAGENTA="\033[0;35m"     # Tool Header Color
BLUE="\033[0;34m"        # Progress
LIGHT_BLUE="\033[1;34m"  # Light Blue Color for Progress
PURPLE="\033[0;35m"      # Purple
LIGHT_CYAN="\033[1;36m"  # Light Cyan
ORANGE="\033[0;33m"      # Orange for highlights
LIGHT_GREEN="\033[1;32m" # Light Green for success

start_time=$(date +%s)

clear
echo -e "${MAGENTA}=========================================================================${RESET}"

echo -e  "██╗     ██╗████████╗ ██████╗  ██████╗███████╗ █████╗ ███╗   ██╗"
echo -e  "██║     ██║╚══██╔══╝██╔═══██╗██╔════╝██╔════╝██╔══██╗████╗  ██║"
echo -e  "██║     ██║   ██║   ██║   ██║██║     █████╗  ███████║██╔██╗ ██║"
echo -e  "██║     ██║   ██║   ██║V  ██║██║ S   ██╔══╝7 ██╔══██║██║╚██╗██║"
echo -e  "███████╗██║   ██║   ╚██████╔╝╚██████╗███████╗██║  ██║██║ ╚████║"
echo -e  "╚══════╝╚═╝   ╚═╝    ╚═════╝  ╚═════╝╚══════╝╚═╝  ╚═╝╚═╝  ╚═══╝" 

echo -e "${MAGENTA}=========================================================================${RESET}"

current_date_time=$(date '+%Y-%m-%d %H:%M:%S') # Get current date and time
echo -e "${LIGHT_BLUE}=============================================================="
echo -e "${LIGHT_CYAN}               LitOcean Subdomain Enumeration Tool             "
echo -e "${LIGHT_BLUE}=============================================================="
echo -e " "
echo -e "${WHITE}Author:${RESET} ${PURPLE}Biswajeet Ray${RESET}"
echo -e "${WHITE}Date:${RESET} ${CYAN}$current_date_time${RESET}"
echo -e "${LIGHT_BLUE}=============================================================="
echo -e " "

echo -e "${GREEN}[*] Starting LitOcean Subdomain Enumeration Tool...${RESET}"
echo -e "\n"

read -p "Enter the Target Domain (e.g. example.com): " domain

if [ -z "$domain" ]; then
    echo -e "${RED}[!] No domain provided. Exiting...${RESET}"
    exit 1
fi

mkdir -p "$domain/subdomains"

fetch_subdomains() {
    tool_name=$1
    command=$2
    echo -e "${LIGHT_GREEN}[+] Fetching subdomains from $tool_name...${RESET}"
    
    eval "$command" &>/dev/null
    echo -e "${LIGHT_CYAN}[*] Subdomains from $tool_name collected successfully.${RESET}"
    eval "$command" >> "$domain/subdomains/$tool_name.txt"
}

fetch_json_subdomains() {
    tool_name=$1
    api_url=$2
    jq_filter=$3
    regex_pattern=$4
    echo -e "${LIGHT_GREEN}[+] Fetching subdomains from $tool_name API...${RESET}"
    response=$(curl -s "$api_url")
    
    if echo "$response" | jq . > /dev/null 2>&1; then
        echo "$response" | jq "$jq_filter" | grep -Po "$regex_pattern" >> "$domain/subdomains/$tool_name.txt"
        echo -e "${LIGHT_CYAN}[*] Subdomains from $tool_name API collected successfully.${RESET}"
    else
        echo -e "${RED}[!] Invalid or Non-JSON Response from $tool_name API${RESET}"
    fi
}

echo -e "${BLUE}[+] Running Subdomain Enumeration Tools...${RESET}"

fetch_subdomains "Amass" "amass enum -passive -d $domain -o /dev/null"

fetch_subdomains "Assetfinder" "assetfinder --subs-only $domain"

fetch_subdomains "Subfinder" "subfinder -d $domain -silent"

fetch_json_subdomains "crt.sh" "https://crt.sh/?q=%25.$domain&output=json" '.[].name_value' "\w.*$domain"

fetch_subdomains "Archive" "curl -s \"http://web.archive.org/cdx/search/cdx?url=*.$domain/*&output=text&fl=original&collapse=urlkey\" | sed -e 's_https*://__' -e \"s/\/.*//\""

echo -e "${LIGHT_BLUE}[+] Combining results, removing duplicates, and filtering...${RESET}"

cat "$domain/subdomains/"*.txt | sort -u | anew "$domain/all_subdomains.txt"

rm -rf "$domain/subdomains"

end_time=$(date +%s)
execution_time=$((end_time - start_time))

if ((execution_time < 60)); then
    echo -e "${GREEN}[+] Subdomain Enumeration Finished in $execution_time seconds${RESET}"
else
    minutes=$((execution_time / 60))
    seconds=$((execution_time % 60))
    echo -e "${GREEN}[+] Subdomain Enumeration Finished in $minutes minutes and $seconds seconds${RESET}"
fi

if [ -f "$domain/all_subdomains.txt" ]; then
    echo -e "${LIGHT_GREEN}[+] Results saved in $domain/all_subdomains.txt${RESET}"
else
    echo -e "${RED}[!] No results found. Please check the domain and try again.${RESET}"
fi
