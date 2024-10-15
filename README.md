# HAProxy montitor and log analyzer


## Routes


* `/stats/frontend/top/:frontend_name` - get top list of IP
* `/stats/frontend` - get rate and request time history for all frontends 
* `/stat/top_ip/:rate` - get list of IPs that exceeed the rate




## hamon-ipset-loader

`hamon-ipset-loader` loads the toplist of IPs from hamon to ipset `hamon-blocked`