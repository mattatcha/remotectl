# remotectl

based on https://github.com/crosbymichael/slex
uses https://github.com/MattAitchison/envconfig
uses https://gist.github.com/MattAitchison/ea7664793653b8fa88dd

## usage

remotectl <flags> <query> [--] <cmd>

### examples

	# run uptime on all hosts
	remotectl -- uptime

	# run uptime on all hosts tagged foo
	remotectl foo uptime

	# list all hosts
	remotectl -l

	# add hosts tagged "app" into /etc/hosts
	remotectl -l app >> /etc/hosts

	# show configuration using a profile (no query or cmd)
	remotectl -p myprofile

### flags

--list, -l
lists selected ips and names. /etc/hosts friendly output.
ignores <cmd>. includes name alias with provider as tld.

--profile, -p <filepath>
sources a bash profile to load a config.

--version, -V
--help, -h

### config

REMOTECTL_IDENTITY
	private key file. openssh default


REMOTECTL_USER
	user to connect as. defaults to current user.

REMOTECTL_PORT
	ssh port to use. all ports must be the same on hosts for a run.

REMOTECTL_PREFIX
	prefix template for host output. uses go templates with values:
		.Name - name of host
		.IP - ip of host
		.Index - numerical index of host in run
		.Group - group index (if groups mode is used)
	a function is provided `cc` that adds a color cycle to host
output hashed on an input string. default:
{{cc .Name}}{{.Name}}:

REMOTECTL_NAMESPACE
	only use hosts from provider with name starting with this prefix.
	namespace prefix is removed from names before being used.

REMOTECTL_PROFILE
	default profile to load if not provided.

REMOTECTL_PROVIDER
	name or comma-sep list of provider modules to use for selecting hosts
	examples: do,ec2,consul



## query parsing

Providers may implement different query string semantics. However, there is an ideal convention set forth by builtin providers:

	Queries are treated as "tag" lookups first. If no hosts exist with
	this tag, query is treated as a "name" lookup. This optimizes for
	primary use case of multiple hosts, but still works to reference
	individual hosts.

	Key-value attributes can be provided in the form `<key>:<value>`.
	Providers implement more detailed semantics around this convention.

## builtin providers

### ec2

AWS_ACCESS_KEY_ID
AWS_SECRET_ACCESS_KEY
EC2_FILTER
apply ec2 filters
EC2_IPDOMAIN public / private
which ip to use (ie, private when using bastion/gateway)


### digital ocean

DO_ACCESS_TOKEN

REMOTECTL_NAMESPACE
limit DO hosts to this named starting with prefix.
allows queries within a namespace since DO has no tags/attributes/vpcs. see REMOTECTL_NAMESPACE

# DEV PLAN

## MVP, 0.1.0

Basic functionality supporting only these flags:
	--version
	--list
	--help
	--profile

And these config options:
	REMOTECTL_PORT
	REMOTECTL_IDENTITY
	REMOTECTL_USER
	REMOTECTL_PROVIDER

And these providers:
	digitalocean

Silent release on Github.

## First rev, 0.2.0

Add these flags:
	--verbose
	--env
	--local
	--first

These config options:
	REMOTECTL_BASHENV
	REMOTECTL_AGENT
	REMOTECTL_PREFIX
	REMOTECTL_NAMESPACE

These providers:
	ec2

Use with clients. Public mentions on Twitter.

REMOTECTL_SHELLENV
	file to copy and source on remote side

REMOTECTL_AGENT
	boolean to forward agent. defaults to true?

	# broadcast docker image to hosts tagged "app" via stdin
	docker save myimage | remotectl app docker load

	# wait for all "app" hosts to be ssh reachable
	remotectl --wait app

	--verbose, -v
	shows status output more like capistrano or fabric.
	otherwise, focuses only on output of remote hosts.

--local, -L
doesn't use ssh, but shells out locally to run <cmd>.
replaces string "{}" in <cmd> with each host ip.
used for rsync, etc. also allows using system ssh.

--first, -1
	limits list of hosts returned to just the first host.
	can be combined with --random to pick random host.

--env, -e <key=value>
sets environment variable to be set in remote shell.

## Public release, 0.3.0

Add these features:
	- Bastion/gateway
	- Modes
	- Random flag
  - Wait flag

Screencast!

REMOTECTL_GATEWAY
selector to use as bastion host.
optional prefix "<user>@" to specify bastion user.

REMOTECTL_MODE
	execution mode options (based on ruby sshkit):
		parallel
			runs commands in parallel. default.
		sequence[:<wait-duration>]
			runs in sequence with option duration between.
			ex: REMOTECTL_MODE=sequence:2s
		groups:<limit>[:<wait-duration>]
			runs in parallel on <limit> hosts, grouped sequential.
ex: REMOTECTL_MODE=group:4:2s


--random, -r
	randomizes the list of hosts returned from query from any
	sorting or bias of the provider.

--wait, -w [timeout-duration]
	polls hosts until they're all ready to receive ssh commands.
	can be used with or without <cmd>. <cmd> is only performed
	when all hosts are ready.

## License

MIT
