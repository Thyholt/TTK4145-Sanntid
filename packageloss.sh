
# iptables -A for adding rule
# iptables -D for deleting rule

#sudo iptables -A INPUT -m statistic --mode random --probability 0.9 -j DROP
sudo iptables -D INPUT -m statistic --mode random --probability 0.9 -j DROP
