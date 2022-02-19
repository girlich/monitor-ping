# monitor-ping
Ping many hosts in a local network. Reads a json structure from stdin and prints this json structure
enriched with the ping answers to stdout.
Optionally acts as a prometheus exporter and queries the hosts from stdin every time a prometheus
scraper comes along.
