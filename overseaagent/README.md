#overseaagent

## overseagaent 是一个查询google play & app store 购买服务的代理.

### 配置
- 配置文件格式为.json,示例参看 ***cmd/agent/config.json***
- 配置文件路径通过命令行传入 eg: agent -conf="configure_file_path.json"
- 配置文件简介:

        {
          "stores": {
            "googleplay": {
              "power_on":        true,                                                             //是否启用google play模块
              "http_proxy":      true,                                                             //是否启用代理                                                 
              "client_id":       "",                                                               //oauth2 client_id
              "client_secret":   "",                                                               //oauth2 client_secret
              "refresh_token":   "",                                                               //oauth2 refresh_token
              "url":             "https://www.googleapis.com/androidpublisher/v2/applications"     //google play 查询url
            },
            "appstore": {
              "power_on":    true,                                                                 //是否启用app store模块  
              "http_proxy":  true,                                                                 //是否启用代理
              "debug":       true,                                                                 //是否sandbox模式                                  
              "debug_url":   "https://sandbox.itunes.apple.com/verifyReceipt",                     //sandbox模式url
              "release_url": "https://buy.itunes.apple.com/verifyReceipt"                          //product模式url
            }
          },
          "host": {
            "port":"8080"                                                                          //本应用监听端口
          },
          "proxy": {          
            "all_on":   true,                                                                      //是否启用全局代理
            "address": "http://192.168.6.72:1080"                                                  //代理地址
        
          }
        
        }
                   
### 运行
       agent.exe [-conf="configure_file_path.json"] 
             
### 注意事项 
1. 如果启用了全局代理模式, 则分商店设置无效.
2. 如果某个商店模块被启用, 则相应配置必须设置完备,否则程序不可启动.