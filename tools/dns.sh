#!/bin/bash

# ANSI颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
NC='\033[0m' # No Color

# 配置文件
CONFIG_FILE="$HOME/.dns_config"
BACKUP_DIR="$HOME/.dns_backup"

# 创建备份目录
mkdir -p "$BACKUP_DIR"

# 打印分隔线
print_separator() {
    echo -e "${BLUE}════════════════════════════════════════════════════════════${NC}"
}

# 显示进度条
show_progress() {
    local duration=$1
    local steps=20
    local step_duration=$(echo "scale=2; $duration/$steps" | bc)
    
    echo -ne "${PURPLE}["
    for ((i=0; i<steps; i++)); do
        echo -ne "="
        sleep $step_duration
    done
    echo -e "]${NC}"
}

# 备份当前DNS配置
backup_dns_config() {
    local timestamp=$(date +"%Y%m%d_%H%M%S")
    local backup_file="$BACKUP_DIR/dns_config_$timestamp.txt"
    networksetup -getdnsservers Wi-Fi > "$backup_file"
    echo -e "${GREEN}✓ DNS configuration backed up to $backup_file${NC}"
}

# 获取所有网络接口
get_network_interfaces() {
    networksetup -listallnetworkservices | grep -v "An asterisk (*) denotes that a network service is disabled"
}

# 获取当前DNS配置
get_dns_servers() {
    scutil --dns | grep 'nameserver' | awk '{print $3}'
}

# 检查网络连接状态
check_network_status() {
    echo -e "\n${BLUE}Network Status:${NC}"
    if ping -c 1 8.8.8.8 &> /dev/null; then
        echo -e "  ${GREEN}✓ Internet connection is active${NC}"
    else
        echo -e "  ${RED}✗ No internet connection${NC}"
    fi
}

# 检查DNS配置
check_dns() {
    local dns_servers
    dns_servers=$(get_dns_servers)
    local has_other_dns=false
    
    print_separator
    echo -e "${BLUE}🔍 DNS Configuration Check${NC}"
    print_separator
    
    # 显示网络接口信息
    echo -e "\n${BLUE}Available Network Interfaces:${NC}\n"
    while IFS= read -r interface; do
        echo -e "  ${PURPLE}• $interface${NC}"
    done <<< "$(get_network_interfaces)"
    
    echo -e "\n${BLUE}Current DNS Servers:${NC}\n"
    
    while IFS= read -r server; do
        if [ "$server" != "127.0.0.1" ]; then
            has_other_dns=true
            echo -e "  ${RED}✗ $server${NC}"
        else
            echo -e "  ${GREEN}✓ $server${NC}"
        fi
    done <<< "$dns_servers"
    
    if [ "$has_other_dns" = true ]; then
        echo -e "\n${YELLOW}⚠️  Non-local DNS servers detected.${NC}"
        echo -e "${YELLOW}🔄 Fixing configuration...${NC}\n"
        
        # 备份当前配置
        backup_dns_config
        
        # 更新DNS配置
        sudo networksetup -setdnsservers Wi-Fi 127.0.0.1
        
        # 等待DNS配置更新
        echo -e "${YELLOW}⏳ Waiting for DNS configuration to update...${NC}"
        show_progress 2
        
        # 显示修复后的DNS配置
        echo -e "${GREEN}✓ DNS configuration updated successfully${NC}"
        echo -e "\n${BLUE}Updated DNS Servers:${NC}\n"
        local updated_servers=$(get_dns_servers)
        local update_success=true
        
        while IFS= read -r server; do
            if [ "$server" = "127.0.0.1" ]; then
                echo -e "  ${GREEN}✓ $server${NC}"
            else
                echo -e "  ${RED}✗ $server${NC}"
                update_success=false
            fi
        done <<< "$updated_servers"
        
        if [ "$update_success" = false ]; then
            echo -e "\n${YELLOW}⚠️  DNS configuration update failed.${NC}"
            echo -e "${YELLOW}🔄 Retrying...${NC}"
            sleep 1
            sudo networksetup -setdnsservers Wi-Fi 127.0.0.1
            sleep 1
            check_dns
            return
        fi
        
        # 检查网络连接
        check_network_status
    else
        echo -e "\n${GREEN}✓ DNS configuration is correct${NC}"
        check_network_status
    fi
    
    print_separator
}

# 主菜单
show_menu() {
    echo -e "\n${BLUE}DNS Configuration Manager${NC}"
    echo -e "1. Check and fix DNS configuration"
    echo -e "2. View backup history"
    echo -e "3. Restore from backup"
    echo -e "4. Exit"
    echo -ne "\n${PURPLE}Select an option: ${NC}"
    read choice
    
    case $choice in
        1) check_dns ;;
        2) ls -l "$BACKUP_DIR" ;;
        3) 
            echo -e "\n${BLUE}Available backups:${NC}"
            ls -l "$BACKUP_DIR"
            echo -ne "\n${PURPLE}Enter backup file name: ${NC}"
            read backup_file
            if [ -f "$BACKUP_DIR/$backup_file" ]; then
                sudo networksetup -setdnsservers Wi-Fi $(cat "$BACKUP_DIR/$backup_file")
                echo -e "${GREEN}✓ DNS configuration restored from $backup_file${NC}"
            else
                echo -e "${RED}✗ Backup file not found${NC}"
            fi
            ;;
        4) exit 0 ;;
        *) echo -e "${RED}Invalid option${NC}" ;;
    esac
}

# 执行主菜单
while true; do
    show_menu
done