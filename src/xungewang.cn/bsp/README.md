## BSP Server Startup
* copy `src/xungewang.cn/bsp/config/app.template.yaml` to some folder, and get it 
renamed as `app.yaml`. For simplicity, let's assume the folder containing that app.yaml 
as `$CONFIG_DIR`.
* start the app with `/path/to/bsp -c $CONFIG_DIR`. The path to the `bsp` 
program is where you have the program compiled. Typically, it's in the 
`bin` folder compiled with `go build xungewang.cn/bsp` command.

###### For building the program
* make sure the project is in $GOPATH, and run following commands: 

```bash
cd ./bin
go build xungewang.cn/bsp
```

 After that, `bsp` becomes available.
 

## BSP Server Restart
After starting your server you can make some changes, build, and send SIGHUP to the running 
process and it will finish handling any outstanding requests and serve all new incoming ones 
with the new binary. `kill -HUP $PID_OF_BSP`.

For finding PID of `bsp`, you may find `ps -ef | grep bsp | grep -v grep | awk '{print $2}' | xargs kill -HUP ` useful.