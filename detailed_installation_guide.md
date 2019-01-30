# Installation Guide:

This guide will go over how to set everything up so that you can get started with developing on Confbase!

## Setting up your Go environment: 
For me personally, I did the following: 

1. Install Go and openssh-server
    - `sudo snap install --classic go` 
        - I used snap, but feel free to use whatever package manager you prefer.
    - `sudo apt install openssh-server`
        - You'll want to start this up when you log into or turn on your computer
            - `sudo systemctl start sshd`
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
        cd $GOPATH/src/github.com/Confbase/cfgd && go install
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
2. Setting up an initial base for finishing up installation
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
    
