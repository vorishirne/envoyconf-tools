# envoke

Envoy is a proxy, really awesome and we are 
devs who often use it, face errors and struggle to debug it, 
when envoy config's source is an external endpoint.

Now, convert the external endpoint config to static configs,
which is what is the requirement for 

1. better visibility
2. better reproducibility of scenarios
3. better experimentation

# How to use

First get the config dump, from the envoy endpoint on admin port.

Save it in a file.

Call the Envoke binary as follows on the config file.

# Example

1. The config dump file is saved with name `envoy-config.json` .
Remember: the file extension should be .json
2. Call binary as `./envoke envoy-config`

3. then load the envoy config as
`cd envoy-config && envoy -c bootstrap.yaml`