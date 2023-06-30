# Machine evidence database

This database is created to contain list of machines on single place.

Machines have given attributes and can have optional labels.

Service can run in readonly mode or in edit mode where you can update your machines using endpoints.

The service is written in GoLang

Client library for GO can be found [here](https://github.com/westwardharbor0/go-server-evidence).

# Running locally
To install all needed dependencies run `go get`.  
To view the help with all the arguments run `go run . -h`

# Configuration 

Configuration to set up the service attributes, machines and machines activity checks.

Example of service configuration:
```yaml
config:
  api:
    host: "localhost"                   # Host we are running on
    port: 8080                          # Post we are running on
    auth: false                         # Toggle for authentication using bearer token
    bearer_token: "random-bearer-token" # Bearer token to be used for authentication
  machines:
    readonly: false         # Toggle for enabling editing of machines and updating machine YAML file
    file: "./machines.yaml" # File for storing machines. Get updated if not in readonly mode
    dump_interval: 5m       # Interval for storing changes done using API
  activity_check:
    check_path: "/health"                    # Endpoint we check on each machine for activity
    check_interval: 5m                       # Interval we run checks in
    check_protocol: "https"
    retries: 3                                    # Amount of times we tolerate failed check
    alert_endpoint: "alerting-hostname.io/report" # Endpoint we send request too if detect inactive machine
```

# Machines

Definitions of machines that will be served and used in the service

Example of machines file:
```yaml
machines:
    example-hostname:              # Repeatable machine record
      hostname: "example-hostname" # Hostname of machine
      active: true                 # If machine is active or inactive due to maintenance for example
      ipv4: "1.1.1.1"              # IPV4 address of machine
      ipv6: "no::body:uses"        # IPV6 address of machine
      labels:                      # Optional machine labels
        group: "storage"
        team: "infra-1"
```

# Endpoints

Status endpoints:
```
GET: /metrics       -- Return prometheus metrics
GET: /status        -- Status page with basic overview
GET: /health        -- Health status if service is OK
```

Machines endpoints (all responses in JSON format):
```
-- Listing endpoints -- 
GET: /machines                 -- Returns list of all machines
GET: /machines/<field>/<value> -- Returns sub set of filtered machines based on url

-- Edit endpoints -- 
DELETE: /machines/<hostname>   -- Deletes machines from service
PUT: /machines                 -- Adds machine to service
POST: /machines                -- Updates machines in service
```

Expected JSON for updating or adding machines: 
```json
{
  "<machine-hostname>": {
    "hostname": "<machine-hostname>",
    "active": true,
    "ipv4": "<ipv4>",
    "ipv6": "<ipv6>",
    "labels": {
      "label_name": "<label_val>"
    }
  }
}
```

# Deploying 

To deploy this service you will need to prepare your config and machines file.

These files can be either baked into a docker image or mounted afterwards.

## TBD

1) Create a public docker image to be pulled and used
2) Document API endpoints better
