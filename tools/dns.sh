sudo networksetup -setdnsservers Wi-Fi 8.8.8.8
sudo networksetup -setdnsservers Wi-Fi 127.0.0.1
scutil --dns | grep 'nameserver'