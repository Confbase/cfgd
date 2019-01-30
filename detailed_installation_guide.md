# Installation Guide:

This guide will go over how to set everything up so that you can get started with developing on Confbase!

## Setting up your Go environment: 
For me personally, I did the following: 

1. Install Go, git and openssh-server
    - `sudo snap install --classic go` 
        - I used snap, but feel free to use whatever package manager you prefer.
    - `sudo apt install openssh-server`
        - You'll want to start this up when you log into or turn on your computer
            - `sudo systemctl start sshd`
    - `sudo apt install git`
        - You'll want to setup SSH keys for git now if you haven't already. [Here's documentation on doing that](https://help.github.com/articles/generating-a-new-ssh-key-and-adding-it-to-the-ssh-agent/). 
2. Create a go workspace
    - I created `~/workspace/go` and `~/workspace/go/bin`
3. Make sure your path variable and env variables are set for Go
    - I added the following to my `~/.bashrc` file: 
    ```
    export GOPATH=~/workspace/go
    export GOBIN=$GOPATH/bin
    export PATH=$PATH:/snap/bin:$GOBIN
    ```
4. You can use whatever editor you prefer, but I opted to use VSCode and the Go plugin.
5. It would be a good idea to install [delve](https://github.com/go-delve/delve) while you're at it
    - `go get -u github.com/go-delve/delve/cmd/dlv`

## Setting up cfg and cfgd

1. Install the go packages
    - There's 3 main packages, first run: 
        ```
        go get github.com/Confbase/cfgd
        go get github.com/Confbase/cfg
        go get github.com/Confbase/schema
        ```
    - Now, run the following: 
        ```
        cd $GOPATH/src/github.com/Confbase/cfgd && go install && go install ./...
        cd $GOPATH/src/github.com/Confbase/cfg && go install
        cd $GOPATH/src/github.com/Confbase/schema && go install
        ```
    - Make sure they copied to disk-- you should get a similar output to this when running the tree command from your `GOPATH` 
        
        ```
        drake@element:~/workspace/go$ tree -L 3
        .
        |-- bin
        |   |-- cfg
        |   |-- cfgd
        |   |-- cfgsnap
        |   |-- dlv
        |   `-- schema
        |-- pkg
        |   `-- linux_amd64
        |       `-- github.com
        `-- src
            |-- github.com
            |   |-- Confbase
            |   |-- clbanning
            |   |-- fsnotify
            |   |-- go-delve
            |   |-- go-redis
            |   |-- hashicorp
            |   |-- magiconair
            |   |-- mitchellh
            |   |-- naoina
            |   |-- pelletier
            |   |-- sirupsen
            |   `-- spf13
            |-- golang.org
            |   `-- x
            `-- gopkg.in
                `-- yaml.v2

        22 directories, 5 files
        ```
2. Setting up an initial base
    - First, we need to set up a folder for our new base
        ```
        cd ~/workspace && mkdir test_base && cd test_base
        cfg init
        printf "%s\n%s" "host: localhost" "port: 5000" >> config.yaml
        cfg mark -t config config.yaml
        ```
    - You should now be able to run `cfg ls` and get similar output to this: 
        ```
        drake@element:~/workspace/test_base$ cfg ls
        ## master
        templates
        config: config.yaml

        instances

        singletons
        .gitignore
        ```
    - Add a remote server to your base
        -   This command should work, but you won't be able to push just yet.
            ```
            cfg remote add origin git@localhost:/srv/git/confbase/test_base
            ```
    - Make sure you've started sshd with `sudo systemctl start sshd`
    - Set up your ssh whitelist to only allow incoming connections from localhost. The top answer on [this stackoverflow post](https://askubuntu.com/questions/179325/accepting-ssh-connections-only-from-localhost) should get you covered.
3. Setting up the git user and backend
    - For this step, we'll need to create a new user, named `git`. This will obviously cause problems if you named yourself git for some reason when setting up your computer. I also have only run / tested this in linux, so YMMV on other platforms. This will largely follow along with [this article](https://git-scm.com/book/en/v2/Git-on-the-Server-Setting-Up-the-Server).
    - Run these commands
        ```
        $ sudo adduser git
        $ su git
        $ cd 
        $ mkdir .ssh && chmod 700 .ssh
        $ touch .ssh/authorized_keys && chmod 600 .ssh/authorized_keys
        ```
    - Locate public key you generated for your github account, and copy it into your clipboard.
    - Paste the whole public key into `~/.ssh/authorized_keys`.
    - Run the following commands: 
        ```
        sudo su
        mkdir -p /srv/git/confbase/test_base
        chown -R git /srv/git
        cd /srv/git/confbase/test_base 
        sudo -u git git init --bare
        cp $GOPATH/src/github.com/Confbase/cfgd/post-receive /srv/git/confbase/test_base/hooks # you might need to do this as sudo
        chown git /srv/git/cofbase/test_base/hooks/post-receive
        chmod u+x post-receive
        ```
4. The first successful push (hopefully)
    - If you're still the git user, exit out of that and go back to your normal account.
    - Run cfgd in the background 
        ```
        cfgd &
        ```
        - Alternatively, just run `cfgd` in another terminal.
    - Go back to the original test_base
        ```
        cd ~/workspace/test_base
        ```
    - Run the following command. You'll know everything works if your output matches mine.
        ```
        drake@element:~/workspace/test_base$ cfg new config yet_another_test_config.yaml && cfg push origin master
        remote: Already on 'master'
        To localhost:/srv/git/confbase/test_base
        2e0da3b..5a93411  master -> master
        ```

Congrats-- if your final output matches, you're all set up!