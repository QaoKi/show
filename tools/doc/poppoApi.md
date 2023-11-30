
# 一、得到Access Token
得到access token用于后续接口请求

## 接口地址
http(s)://<ip>:<port>/user/token
## 请求方式
POST
## 请求格式
application/json
## 请求参数
|  参数名 |  说明 | 类型  |
| ------------ | ------------ | ------------ |
|  k |  appKey，游戏中心索取  |  String |
|  t   | 时间戳（秒）  |  Number |
|  id   | IdToken  |  String |
|  sign   | 签名  |  String |

## 响应
|  参数名 |  说明 | 类型  |
| ------------ | ------------ | ------------ |
|  code |  返回码，参考全局返回码  |  Number |
|  message   | 返回信息  |  String |
|  data   | 返回体  |  Object |
|  data.access_token   | Access Token，用于后续接口请求，请求后先前access token失效  |  String |
|  data.expires_in   | Access Token有效期（秒）  |  Number |
|  data.refresh_token   | Refresh Token，用于刷新令牌  |  String |
|  data.refresh_expires_in   | Refresh Token有效期（秒）  |  Number |

## 签名生成
请求参数按key ascii正序后与对应值、appSecret（游戏中心索取）拼接后sha512，参考node代码：

![](https://www.showdoc.com.cn/server/api/attachment/visitFile?sign=9e90d6c86658d5e2f95449bc9afaaa20&file=file.png)


# 二、刷新令牌

## 接口地址
http(s)://<ip>:<port>/user/refresh_token
## 请求方式
POST
## 请求格式
application/json
## 请求参数
|  参数名 |  说明 | 类型  |
| ------------ | ------------ | ------------ |
|  k |  appKey，游戏中心索取  |  String |
|  refresh_token   | 刷新令牌  |  String |

## 响应
|  参数名 |  说明 | 类型  |
| ------------ | ------------ | ------------ |
|  code |  返回码，参考全局返回码  |  Number |
|  message   | 返回信息  |  String |
|  data   | 返回体  |  Object |
|  data.access_token   | Access Token，用于后续接口请求，请求后先前access token失效  |  String |
|  data.expires_in   | Access Token有效期（秒）  |  Number |
|  data.refresh_token   | Refresh Token，用于刷新令牌  |  String |
|  data.refresh_expires_in   | Refresh Token有效期（秒）  |  Number |


# 三、查询用户余额

## 接口地址
http(s)://<ip>:<port>/user/balance?access_token=ACCESS_TOKEN
## 请求方式
GET

## 响应
|  参数名 |  说明 | 类型  |
| ------------ | ------------ | ------------ |
|  code |  返回码，参考全局返回码  |  Number |
|  message   | 返回信息  |  String |
|  data   | 余额  |  Number |


# 四、加钱

## 接口地址
http(s)://<ip>:<port>/user/money_add?access_token=ACCESS_TOKEN
## 请求方式
POST
## 请求格式
application/json
## 请求参数
|  参数名 |  说明 | 类型  |
| ------------ | ------------ | ------------ |
|  money |  要加的钱，正整数  |  Number |

## 响应
|  参数名 |  说明 | 类型  |
| ------------ | ------------ | ------------ |
|  code |  返回码，参考全局返回码  |  Number |
|  message   | 返回信息  |  String |


# 五、扣钱

## 接口地址
http(s)://<ip>:<port>/user/money_deduct?access_token=ACCESS_TOKEN
## 请求方式
POST
## 请求格式
application/json
## 请求参数
|  参数名 |  说明 | 类型  |
| ------------ | ------------ | ------------ |
|  money |  要扣的钱，正整数  |  Number |

## 响应
|  参数名 |  说明 | 类型  |
| ------------ | ------------ | ------------ |
|  code |  返回码，参考全局返回码  |  Number |
|  message   | 返回信息  |  String |


##全局返回码说明
|  返回码 |  说明  |
| ------------ | ------------ |
|  0 |  请求成功 |
|  40001  |  错误的客户端 |
|  41002  |  appKey为空 |
|  41003  |  时间戳为空 |
|  41004  |  IdToken为空 |
|  41005  |  签名为空 |
|  41006  |  无效签名 |
|  41007  |  无效时间戳 |
|  41008  |  无效IdToken |
|  41009  |  无效access token |
|  41010  |  无效用户 |
|  41011  |  无效金额 |
|  41012  |  金额更新失败 |
|  41013  |  access token生成失败 |
|  41014  |  refresh token为空 |
|  41015  |  无效refresh token |