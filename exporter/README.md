
# 一个将数据从mongodb导出的内部工具.

## 导出原则如下:
1. 按用户id分组
2. 按产品分组
3. 按产品版本分组
4. 按渠道分组

## 需要导出的数据有:
 * 某用户的所有签名文件
 * 某产品的icon
 * 某产品按版本分类的原始二进制文件
 * 某产品按版本按渠道分类的2次打包的二进制文件
 
## 目录结构:

假设根目录为base
	
	base
	   /user
	   		/certs
	   			/example.cert	   			
	   		/apps
	   			/app1(1-N)
	   				/icon
	   					/icon.jpeg	
	   				/versions
	   					/version1(1-N)
	   						/origin
	   							/unsigned.apk
	   						/signed
	   							/signed.apk
	   			
			


