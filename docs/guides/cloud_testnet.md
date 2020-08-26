<!--
order: 4
-->

# Deploy Testnet on Cloud Provider

Learn how to deploy testnet to different cloud providers. {synopsis}

## Pre-requisite Readings

- [Testnet Quickstart](./../quickstart/testnet.md) {prereq}

## Digital Ocean

### Account Setup

Head over to [Digital Ocean](https://www.digitalocean.com/) and create an account.

DigitalOcean will want a public key that it can place on any Droplets we start, so that we can access them with a key that we know only we have.

Let's create an SSH keypair now using `ssh-keygen -t rsa -b 4096`

This will ask you for a file where you want to save the key which you can call something like - `digital-ocean-key`.

It'll also ask for a passphrase - feel free to set one if you wish or you could leave it empty. If you created it in the same folder as we've been working out of, you'll see two files - one called `digital-ocean-key` and one called `digital-ocean-key.pub` - these are respectively your private and public keys.

In your DigitalOcean account, on the bottom left hand side, there is a link for `'Security'`. Follow this link, and the next page will have an option to add an SSH key

Click `'Add an SSH key'` and you'll be presented with a dialog to enter your key. Simply copy the contents of your `digital-ocean-key.pub` into the large text box (you can get the contents printed to the terminal with `cat digital-ocean-key.pub`).

### Create Droplet

Once you've added your SSH key. click on the `'Droplets'` link on the left, and then on the next page click `'Create Droplet'`.

On this page, you'll be presented with a number of options for configuring your DigitalOcean Droplet, including the distribution, the plan, the size/cost per month, region, and authentication. Feel free to choose whichever settings work best for you.

Under `'Authentication'`, select `'SSH Key'`, and select which keys you would like to use (like the one you created in the last step). You can also name your Droplet if you wish. When you're finished, click `'Create Droplet'` at the bottom.

Wait a minute for your Droplet to start up. It'll appear under the `'Droplets'` panel with a green dot next to it when it is up and ready. At this point, we're ready to connect to it.

### Deploy to Droplet

#### Connect to Droplet

Click on the started Droplet, and you'll see details about it. At the moment, we're interested in the IP address - this is the address that the Droplet is at on the internet.

To access it, we'll need to connect to it using our previously created private key. From the same folder as that private key, run:

```bash
ssh -i digital-ocean-key root@<DROPLET_IP_ADDRESS>
```

Now you are connected to the droplet.

#### Install Ethermint

Clone and build Ethermint in the droplet using `git`:

```bash
go install https://github.com/ChainSafe/ethermint.git
```

Check that the binaries have been successfuly installed:

```bash
ethermintd -h
ethermintcli -h
```

### Copy the Genesis File

To connect the node to the existing testnet, fetch the testnet's `genesis.json` file and copy it into the new droplets config directory (i.e `$HOME/.ethermintd/config/genesis.json`).

To do this ssh into both the testnet droplet and the new node droplet.

On your local machine copy the genesis.json file from the testnet droplet to the new droplet using:

```bash
scp -3 root@<TESTNET_IP_ADDRESS>:$HOME/.ethermintd/config/genesis.json root@<NODE_IP_ADDRESS>:$HOME/.ethermintd/config/genesis.json
```

### Start the Node

Once the genesis file is copied over run `ethermind start` inside the node droplet.

## Next {hide}

Follow [Deploy node to public testnet](./deploy_node_on_public_testnet.md)