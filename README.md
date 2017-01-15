
# Docker Machine driver for Joyent Triton

Creates docker instances on a deployment of Joyent Triton. This driver communicates with CloudAPI and will work in Joyent Public Cloud and private deployments

```bash
docker-machine create -d triton machine0
```


## Installation

The easiest way to install the triton docker-machine driver is to:

```
go install github.com/davefinster/docker-machine-driver-triton

`which docker-machine-driver-triton`
```

binaries also available under [releases tab](https://github.com/davefinster/docker-machine-driver-triton/releases)

## Example Usage

Sign up at https://www.ctl.io and export your credentials into your shell environment or pass as cmdline flags

```bash
#using ENV vars
export TRITON_USERNAME='<username>'
export TRITON_KEY_PATH='./path/to/private/key'
export TRITON_KEY_ID='d7:e0.....'

docker-machine -D create -d "triton" --triton-username <username> --triton-key-id d7:e0:.. --triton-key-path ./path/to/private/key machine0
```

## Options

```bash
docker-machine create -d triton --help
```

 Option Name                                          				| Description                                     | Default Value         			 									  | required 
--------------------------------------------------------------------|-------------------------------------------------|-----------------------------------------------------------------------|----------
 ``--triton-username`` or ``TRITON_USERNAME``   					| Triton Username/Account					 	  | none                 												  | yes   
 ``--triton-key-id`` or ``TRITON_KEY_ID``   						| ID of private key used for authentication 	  | none                 												  | yes       
 ``--triton-key-path`` or ``$TRITON_KEY_PATH``      				| File path to private key file         		  | none         														  | yes 
 ``--triton-endpoint`` or ``$TRITON_ENDPOINT``						| CloudAPI HTTP Endpoint 						  | https://us-sw-1.api.joyent.com   									  | no     
 ``--triton-image`` or ``$TRITON_IMAGE``							| UUID of Image to use for VM                     | 698a8146-d6d9-4352-99fe-6557ebce5661 (Ubuntu 14.04.1 LTS Certified)   | no
 ``--triton-image-ssh-username`` or ``$TRITON_IMAGE_SSH_USERNAME``	| SSH Username to connect with (based on image)   | ubuntu                  											  | no         
 ``--triton-network`` or ``TRITON_NETWORK`` 						| UUID of network to connect the VM to            | f7ed95d3-faaf-43ef-9346-15644403b963 (Joyent-SDC-Public)              | no       
 ``--triton-package`` or ``TRITON_PACKAGE``           				| Name of package to provision the VM against     | k4-general-kvm-3.75G         							              | no       

Each environment variable may be overloaded by its option equivalent at runtime.

## Default Image

The default image is [Ubuntu Certified 16.04.1 LTS (20161221 64-bit)](https://docs.joyent.com/images/linux/ubuntu-certified) 

## Hacking

### Get the sources

```bash
go get github.com/davefinster/docker-machine-driver-triton
cd $GOPATH/src/github.com/davefinster/docker-machine-driver-triton
```

### Test the driver

To test the driver make sure your current build directory has the highest
priority in your ``$PATH`` so that docker-machine can find it.

```
export PATH=$GOPATH/src/github.com/davefinster/docker-machine-driver-triton:$PATH
```

## Related links

- **Docker Machine**: https://docs.docker.com/machine/
- **Contribute**: https://github.com/davefinster/docker-machine-driver-triton
- **Report bugs**: https://github.com/davefinster/docker-machine-driver-triton/issues

## License

Apache 2.0