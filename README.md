# cloudflare-audit-log-scraper

Scrapes Audit Logs From Cloudflare and Streams to Std Out. This application is designed to be deployed to any
Kubernetes cluster that has centralized logging that can process valid JSON.

The application will log any Cloudflare audit logs to std-out and is designed to be collected by any centralised
logging solution such as ELK, Loki, Splunk etc.

# Configuration Options

All paramters are configured via environment variables for ease of configuration with a kubernetes deployment. A
deployment has been chosen to ensure that the pod can easily be scheduled across availability-zones.

`CLOUDFLARE_API_EMAIL`
* Description: An email address with at least read access to the cloudflare organization
* Default: null
* Required: true

`CLOUDFLARE_API_KEY`
* Description: An API key associated with the CLOUDFLARE_API_EMAIL
* Default: null
* Required: true

`CLOUDFLARE_ORGANIZATION_ID`
* Description: The organization for which you walt to collect audit logs
* Default: null
* Required: true

`CLOUDFLARE_LOOK_BACK_INTERVAL`
* Description: How far back to look back in minutes
* Default: 5
* Required: false
