# remotectl

based on https://github.com/crosbymichael/slex
uses https://github.com/MattAitchison/envconfig
uses https://gist.github.com/MattAitchison/ea7664793653b8fa88dd

## Usage

	Usage: remotectl <flags> <query> [--] <cmd>

	Providers:
	[do]

	Environment Vars:
	REMOTECTL_USER=""          # user to connect as
	REMOTECTL_PROVIDER="do"        # comma-sep list of provider modules to use for selecting hosts
	REMOTECTL_NAMESPACE=""         # namespace is a prefix which is matched and removed from hosts
	REMOTECTL_PREFIX="{{.Name}}: " # prefix template for host log output
	REMOTECTL_PORT="22"            # port used to connect to each host
	SSH_AUTH_SOCK=""               # ssh agent socket
	REMOTECTL_IDENTITY=""          # file from which the identity (private key) for public key authentication is read.

	Flags:
	  -help
	    	show this help message
	  -list
	    	lists selected ips and names. /etc/hosts friendly output
	  -profile string
	    	bash profile to source for env config
	  -version
	    	show version

## Examples

	# run uptime on all hosts
	remotectl -- uptime

	# run uptime on all hosts tagged foo
	remotectl foo uptime

	# list all hosts
	remotectl -list

	# add hosts tagged "app" into /etc/hosts
	remotectl -list app >> /etc/hosts

	# show configuration using a profile (no query or cmd)
	remotectl -profile myprofile

## Prefix Template

go template prefix for host output.

- .Name - name of host
- .IP - ip of host
- .Index - numerical index of host in run
- .Group - group index (if groups mode is used)

## Query Parsing

Providers may implement different query string semantics. However, there is an ideal convention set forth by builtin providers:

	Queries are treated as "tag" lookups first. If no hosts exist with
	this tag, query is treated as a "name" lookup. This optimizes for
	primary use case of multiple hosts, but still works to reference
	individual hosts.

	Key-value attributes can be provided in the form `<key>:<value>`.
	Providers implement more detailed semantics around this convention.

## Builtin Providers

### Digital Ocean

#### Env
 - `DO_ACCESS_TOKEN`
 - `REMOTECTL_NAMESPACE`
will limit to DO hosts prefixed with namespace.
Allows queries within a namespace since DO has no tags/attributes/vpcs.

## Roadmap

### First rev, 0.2.0

Use with clients. Public mentions on Twitter.

#### Add these flags:
* `-verbose`, `-v`
	shows status output more like capistrano or fabric.
	otherwise, focuses only on output of remote hosts.

* `-env`, `-e <key=value>`
sets environment variable to be set in remote shell.

* `-local`, `-L`
	doesn't use ssh, but shells out locally to run <cmd>.
	replaces string "{}" in <cmd> with each host ip.
	used for rsync, etc. also allows using system ssh.

* `-first`, `-1`
	limits list of hosts returned to just the first host.
	can be combined with --random to pick random host.

#### These config options:
* `REMOTECTL_BASHENV`
*	`REMOTECTL_AGENT`
*	`REMOTECTL_PREFIX`
*	`REMOTECTL_NAMESPACE`
* `REMOTECTL_SHELLENV`
	file to copy and source on remote side
* `REMOTECTL_AGENT`
	boolean to forward agent. defaults to true?

#### These providers:
* ec2
	* `AWS_ACCESS_KEY_ID`
	* `AWS_SECRET_ACCESS_KEY`
	* `EC2_FILTER` apply ec2 filters
	* `EC2_IPDOMAIN` public / private
	which ip to use (ie, private when using bastion/gateway)




#### Feature Examples

	# broadcast docker image to hosts tagged "app" via stdin
	docker save myimage | remotectl app docker load

	# wait for all "app" hosts to be ssh reachable
	remotectl --wait app


### Public release, 0.3.0

Screencast!

#### Add these features:
	- Bastion/gateway
	- Modes
	- Random flag
  - Wait flag

#### Config

* `REMOTECTL_GATEWAY`
selector to use as bastion host.
optional prefix "<user>@" to specify bastion user.

* `REMOTECTL_MODE`
	execution mode options (based on ruby sshkit):

	*	`parallel`
			runs commands in parallel. default.

	*	`sequence[:<wait-duration>]`
			runs in sequence with option duration between.
			ex: REMOTECTL_MODE=sequence:2s

	*	`groups:<limit>[:<wait-duration>]`
			runs in parallel on <limit> hosts, grouped sequential.
			ex: REMOTECTL_MODE=group:4:2s

#### Flags
* `--random`, `-r`
	randomizes the list of hosts returned from query from any
	sorting or bias of the provider.

* `--timeout`, `-t <wait-duration>` global timeout val
* `--wait`, `-w`
	polls hosts until they're all ready to receive ssh commands.
	can be used with or without <cmd>. <cmd> is only performed
	when all hosts are ready.

## License

MIT
