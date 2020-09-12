# gerberos

Simple (log) file watcher and ipset-based banning utility

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
