# gerberos

gerberos scans sources for lines matching regular expressions and containing IPv4 or IPv6 addresses and performs actions on those addresses.
Possible sources are (not necessarily existant) non-directory files and systemd journals.
Found IP addresses can be logged or added to ipsets (`gerberos4` and `gerberos6`) that gerberos will create automatically.

No additional logic (e.g. authentication logic like counting repeated occurrences) is applied. This is to adhere to the [Unix philosophy](https://en.wikipedia.org/wiki/Unix_philosophy), but impacts gerberos' out-of-the-box usefulness for specific use cases when compared to tools like [fail2ban](https://github.com/fail2ban/fail2ban).

## Requirements

- Go 1.15.2 (development only)
- GNU Make 4.3 (development only, optional)
- ipset 6.34
- iptables 1.6.1

## Build

`make build`

## Example configuration file (TOML)

```toml
[rules]
    [rules.apache-fuzzing]
    # Available sources are
    # - ["file", "<path to non-directory file>"]
    # - ["systemd", "<name of systemd service>"]
    source = ["file", "/var/log/apache2/access.log"]
    # "%host%" must appear exactly once in regexp.
    # It will be replaced with a subexpression named
    # "host" matching IPv4 and IPv6 addresses.
    regexp = "%host%.*40(0|8) 0 \"-\" \"-\""
    # Available actions are
    # - ["ban", "<value parsable by time.ParseDuration>"]
    # - ["log"]
    action = ["ban", "1h"]

    [rules.sshd-invalid-user]
    source = ["file", "/var/log/auth.log"]
    regexp = "Invalid user.*%host%"
    action = ["ban", "24h"]
```

## Example systemd service file

```systemd
[Unit]
Description=gerberos
After=network.target

[Service]
Type=simple
Restart=always
RestartSec=5
User=gerberos
WorkingDirectory=/home/gerberos
ExecStart=/home/gerberos/gerberos
CapabilityBoundingSet=CAP_NET_RAW CAP_NET_ADMIN
AmbientCapabilities=CAP_NET_RAW CAP_NET_ADMIN

[Install]
WantedBy=multi-user.target
```
