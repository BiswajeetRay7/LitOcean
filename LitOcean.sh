#!/bin/bash
#### Colors Output
RESET="\033[0m"          # Normal Colour
RED="\033[0;31m"         # Error / Issues
GREEN="\033[0;32m"       # Successful       
BOLD="\033[01;01m"       # Highlight         
WHITE="\033[1;37m"       # Bold Text         
YELLOW="\033[1;33m"      # Warnings and Info 
BLINK="\033[5m"          # Blinking Effect

# Start Timing
start_time=$(date +%s)

# Tool Header with Enhanced Design
tput setaf 5;
tput bold;

echo -e "${YELLOW}==========================================================================${RESET}"
echo -e "${GREEN}"
echo -e  "██╗     ██╗████████╗ ██████╗  ██████╗███████╗ █████╗ ███╗   ██╗"
echo -e  "██║     ██║╚══██╔══╝██╔═══██╗██╔════╝██╔════╝██╔══██╗████╗  ██║"
echo -e  "██║     ██║   ██║   ██║   ██║██║     █████╗  ███████║██╔██╗ ██║"
echo -e  "██║     ██║   ██║   ██║V  ██║██║ S   ██╔══╝7 ██╔══██║██║╚██╗██║"
echo -e  "███████╗██║   ██║   ╚██████╔╝╚██████╗███████╗██║  ██║██║ ╚████║"
echo -e  "╚══════╝╚═╝   ╚═╝    ╚═════╝  ╚═════╝╚══════╝╚═╝  ╚═╝╚═╝  ╚═══╝"
echo -e "${RESET}"
echo -e "${YELLOW}==========================================================================${RESET}"

# Tool Info & Author Credits with Blinking Effect
current_date_time=$(date '+%Y-%m-%d %H:%M:%S') # Get current date and time
echo -e "${WHITE}=============================================================="
echo -e "${WHITE}               LitOcean Subdomain Enumeration Tool             "
echo -e "${WHITE}=============================================================="
echo -e " "
echo -e "${WHITE}#       ******${BLINK}Author: Biswajeet Ray${RESET}${WHITE}******                    #"
echo -e "${WHITE}#       ******${BLINK}Date: $current_date_time${RESET}${WHITE} ******                #"
echo -e "=============================================================="
echo -e " "

# Tool Start Notification
echo -e "${YELLOW}[*] Starting LitOcean Subdomain Enumeration Tool...${RESET}"
echo -e "\n"

# Get the target domain from the user
read -p "Enter the Target Domain (e.g. example.com): " domain

# Check if a domain is provided
if [ -z "$domain" ]; then
    echo -e "${RED}[!] No domain provided. Exiting...${RESET}"
    exit 1
fi

# Create directory for target results
mkdir -p "$domain/subdomains"

# Function to fetch subdomains using different tools
fetch_subdomains() {
    echo -e "${YELLOW}[+] Collecting Subdomains from $1...${RESET}"
    eval "$2" >> "$domain/subdomains/$1.txt"
}

# Function to handle APIs that return JSON
fetch_json_subdomains() {
    echo -e "${YELLOW}[+] Collecting Subdomains from $1 API...${RESET}"
    
    response=$(curl -s "$2")
    
    # Check if the response is valid JSON
    if echo "$response" | jq . > /dev/null 2>&1; then
        echo "$response" | jq "$3" | grep -Po "$4" >> "$domain/subdomains/$1.txt"
    else
        echo -e "${RED}[!] Invalid or Non-JSON Response from $1 API${RESET}"
    fi
}

# Subdomain enumeration from multiple tools

# Amass
fetch_subdomains "Amass" "amass enum -passive -d $domain -o /dev/null"

# Assetfinder
fetch_subdomains "Assetfinder" "assetfinder --subs-only $domain"

# Subfinder
fetch_subdomains "Subfinder" "subfinder -d $domain -silent"

# crt.sh with JSON validation
fetch_json_subdomains "crt.sh" "https://crt.sh/?q=%25.$domain&output=json" '.[].name_value' "\w.*$domain"

# Archive.org
fetch_subdomains "Archive" "curl -s \"http://web.archive.org/cdx/search/cdx?url=*.$domain/*&output=text&fl=original&collapse=urlkey\" | sed -e 's_https*://__' -e \"s/\/.*//\""

# Combine all results, remove duplicates, and filter out empty subdomains
echo -e "${YELLOW}[+] Combining Results, Removing Duplicates, and Filtering...${RESET}"
cat "$domain/subdomains/"*.txt | sort -u | grep -v '^$' | uniq > "$domain/all_subdomains.txt"
rm -rf "$domain/subdomains"

# Script Execution Timing
end_time=$(date +%s)  
execution_time=$((end_time - start_time))  

# Display timing information
if ((execution_time < 60)); then
    echo -e "${GREEN}[+] Subdomain Enumeration Finished in $execution_time seconds${RESET}"
else
    minutes=$((execution_time / 60))  
    seconds=$((execution_time % 60))  
    echo -e "${GREEN}[+] Subdomain Enumeration Finished in $minutes minutes and $seconds seconds${RESET}"
fi

# Final Output
if [ -f "$domain/all_subdomains.txt" ]; then
    echo -e "${GREEN}[+] Results saved in $domain/all_subdomains.txt${RESET}"
else
    echo -e "${RED}[!] No results found. Please check the domain and try again.${RESET}"
fi
