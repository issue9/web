# config


## configOf

在项目正式运行之后，对于配置项的修改应该慎之又慎， 不当的修改可能导致项目运行过程中出错，比如改变了唯一 ID 的生成规则，可能会导致新生成的唯一 ID 与之前的 ID 重复。<br />


| JSON | YAML | XML | TOML | 类型 | 描述 |
|------|------|-----|------|------------------|------------------|
| memoryLimit,omitempty | memoryLimit,omitempty | memoryLimit,attr,omitempty | memoryLimit,omitempty | int64 | 内存限制<br />如果小于等于 0，表示不设置该值。 具体功能可参考[文档](https://pkg.go.dev/runtime/debug#SetMemoryLimit)。除非对该功能非常了解，否则不建议设置该值。<br /> |
| logs,omitempty | logs,omitempty | logs,omitempty | logs,omitempty | [logsConfig](#logsconfig) | 日志系统的配置项<br />如果为空，所有日志输出都将被抛弃。<br /> |
| language,omitempty | language,omitempty | language,attr,omitempty | language,omitempty | string | 指定默认语言<br />服务端的默认语言以及客户端未指定 accept-language 时的默认值。 如果为空，则会尝试当前用户的语言。<br /> |
| http,omitempty | http,omitempty | http,omitempty | http,omitempty | [httpConfig](#httpconfig) | 与 HTTP 请求相关的设置项<br /> |
| timezone,omitempty | timezone,omitempty | timezone,omitempty | timezone,omitempty | string | 时区名称<br />可以是 Asia/Shanghai 等，具体可参考[文档](https://en.wikipedia.org/wiki/List_of_tz_database_time_zones)。<br />为空和 Local(注意大小写) 值都会被初始化本地时间。<br /> |
| cache,omitempty | cache,omitempty | cache,omitempty | cache,omitempty | [cacheConfig](#cacheconfig) | 指定缓存对象<br />如果为空，则会采用内存作为缓存对象。<br /> |
| fileSerializers,omitempty | fileSerializers,omitempty | fileSerializers&gt;fileSerializer,omitempty | fileSerializers,omitempty | string | 指定配置文件的序列化<br />可通过 \[RegisterFileSerializer] 进行添加额外的序列化方法。默认为空，可以有以下可选值：<br />  - yaml 支持 .yaml 和 .yml 两种后缀名的文件<br />  - xml 支持 .xml 后缀名的文件<br />  - json 支持 .json 后缀名的文件<br />  - toml 支持 .toml 后缀名的文件<br />如果为空，表示支持以上所有格式。<br /> |
| compressions,omitempty | compressions,omitempty | compressions&gt;compression,omitempty | compressions,omitempty | [compressConfig](#compressconfig) | 压缩的相关配置<br />如果为空，那么不支持压缩功能。<br /> |
| mimetypes,omitempty | mimetypes,omitempty | mimetypes&gt;mimetype,omitempty | mimetypes,omitempty | [mimetypeConfig](#mimetypeconfig) | 指定可用的 mimetype<br />如果为空，那么将不支持任何格式的内容输出。<br /> |
| idGenerator,omitempty | idGenerator,omitempty | idGenerator,omitempty | idGenerator,omitempty | string | 唯一 ID 生成器<br />该值由 \[RegisterIDGenerator] 注册而来，默认情况下，有以下三个选项：<br />  - date 日期格式，默认值；<br />  - string 普通的字符串；<br />  - number 数值格式；<br />NOTE: 一旦运行在生产环境，就不应该修改此属性，除非能确保新的函数生成的 ID 不与之前生成的 ID 重复。<br /> |
| problemTypePrefix,omitempty | problemTypePrefix,omitempty | problemTypePrefix,omitempty | problemTypePrefix,omitempty | string | Problem 中 type 字段的前缀<br /> |
| onRender,omitempty | onRender,omitempty | onRender,omitempty | onRender,omitempty | string | OnRender 修改渲染结构<br />可通过 \[RegisterOnRender] 进行添加额外的序列化方法。默认为空，可以有以下可选值：<br />  - render200 所有输出都是以 \[server.RenderResponse] 作为返回对象；<br /> |
| registry,omitempty | registry,omitempty | registry,omitempty | registry,omitempty | [registryConfig](#registryconfig) | 指定服务发现和注册中心<br />NOTE: 作为微服务和网关时才会有效果<br /> |
| peer,omitempty | peer,omitempty | peer,omitempty | peer,omitempty | string | 作为微服务时的节点地址<br />NOTE: 作为微服务时才会有效果<br /> |
| mappers,omitempty | mappers,omitempty | mappers&gt;mapper,omitempty | mappers,omitempty | [mapperConfig](#mapperconfig) | 作为微服务网关时的外部请求映射方式<br />NOTE: 作为微服务的网关时才会有效果<br /> |
| user,omitempty | user,omitempty | user,omitempty | user,omitempty | [T](#t) | 用户自定义的配置项<br /> |



## logsConfig




| JSON | YAML | XML | TOML | 类型 | 描述 |
|------|------|-----|------|------------------|------------------|
| location,omitempty | location,omitempty | location,attr,omitempty | location,omitempty | bool | 是否在日志中显示调用位置<br /> |
| created,omitempty | created,omitempty | created,omitempty | created,omitempty | string | 日志显示的时间格式<br />Go 的时间格式字符串，如果为空表示不显示；<br /> |
| levels,omitempty | levels,omitempty | levels&gt;level,omitempty | levels,omitempty | [logs.Level](#logslevel) | 允许开启的通道<br />为空表示采用 \[AllLevels]<br /> |
| std,omitempty | std,omitempty | std,attr,omitempty | std,omitempty | bool | 是否接管标准库的日志<br /> |
| stackError,omitempty | stackError,omitempty | stackError,attr,omitempty | stackError,omitempty | bool | 是否显示错误日志的调用堆栈<br /> |
| handlers | handlers | handlers&gt;handler | handlers | [logHandlerConfig](#loghandlerconfig) | 日志输出对象的配置<br />为空表示 \[NewNopHandler] 返回的对象。<br /> |



## logs.Level

未找到类型的相关文档


## logHandlerConfig




| JSON | YAML | XML | TOML | 类型 | 描述 |
|------|------|-----|------|------------------|------------------|
| levels,omitempty | levels,omitempty | level,omitempty | levels,omitempty | [logs.Level](#logslevel) | 当前 Handler 支持的通道<br />为空表示采用 \[logsConfig.Levels] 的值。<br /> |
| type | type | type,attr | type | string | Handler 的类型<br />可通过 \[RegisterLogsHandler] 方法注册，默认包含了以下几个：<br />  - file 输出至文件<br />  - smtp 邮件发送的日志<br />  - term 输出至终端<br /> |
| args,omitempty | args,omitempty | args&gt;arg,omitempty | args,omitempty | string | 当前日志的初始化参数<br />根据以上的 type 不同而不同：<br />### file: {#hdr-file_}<br />	0: 保存目录；<br />	1: 文件格式，可以包含 Go 的时间格式化字符，以 %i 作为同名文件时的序列号；<br />	2: 文件的最大尺寸，单位 byte；<br />	3: 文件的格式，默认为 text，还可选为 json；<br />### smtp: {#hdr-smtp_}<br />	0: 账号；<br />	1: 密码；<br />	2: 主题；<br />	3: 为 smtp 的主机地址，需要带上端口号；<br />	4: 接收邮件列表；<br />	5: 文件的格式，默认为 text，还可选为 json；<br />### term {#hdr-term}<br />	0: 输出的终端，可以是 stdout 或 stderr；<br />	1-7: Level 以及对应的字符颜色，格式：erro:blue，可用颜色：<br />	 - default 默认；<br />	 - black 黑；<br />	 - red 红；<br />	 - green 绿；<br />	 - yellow 黄；<br />	 - blue 蓝；<br />	 - magenta 洋红；<br />	 - cyan 青；<br />	 - white 白；<br /> |



## httpConfig




| JSON | YAML | XML | TOML | 类型 | 描述 |
|------|------|-----|------|------------------|------------------|
| port,omitempty | port,omitempty | port,attr,omitempty | port,omitempty | string | 端口<br />格式与 \[http.Server.Addr] 相同。可以为空，表示由 \[http.Server] 确定其默认值。<br /> |
| url,omitempty | url,omitempty | url,omitempty | url,omitempty | string | \[web.Router.URL] 的默认前缀<br />如果是非标准端口，应该带上端口号。<br />NOTE: 每个路由可使用 \[web.WithURLDomain] 重新定义该值。<br /> |
| requestID,omitempty | requestID,omitempty | requestID,omitempty | requestID,omitempty | string | x-request-id 的报头名称<br />如果为空，则采用 \[header.XRequestID] 作为默认值。<br /> |
| certificates,omitempty | certificates,omitempty | certificates&gt;certificate,omitempty | certificates,omitempty | [certificateConfig](#certificateconfig) | 网站的域名证书<br />NOTE: 不能同时与 ACME 生效<br /> |
| acme,omitempty | acme,omitempty | acme,omitempty | acme,omitempty | [acmeConfig](#acmeconfig) | ACME 协议的证书<br />NOTE: 不能同时与 Certificates 生效<br /> |
| readTimeout,omitempty | readTimeout,omitempty | readTimeout,attr,omitempty | readTimeout,omitempty | [Duration](#duration) | ReadTimeout 对 \[http.Server.ReadTimeout] 字段<br /> |
| writeTimeout,omitempty | writeTimeout,omitempty | writeTimeout,attr,omitempty | writeTimeout,omitempty | [Duration](#duration) | WriteTimeout 对 \[http.Server.WriteTimeout] 字段<br /> |
| idleTimeout,omitempty | idleTimeout,omitempty | idleTimeout,attr,omitempty | idleTimeout,omitempty | [Duration](#duration) | IdleTimeout 对 \[http.Server.IdleTimeout] 字段<br /> |
| readHeaderTimeout,omitempty | readHeaderTimeout,omitempty | readHeaderTimeout,attr,omitempty | readHeaderTimeout,omitempty | [Duration](#duration) | ReadHeaderTimeout 对 \[http.Server.ReadHeaderTimeout] 字段<br /> |
| maxHeaderBytes,omitempty | maxHeaderBytes,omitempty | maxHeaderBytes,attr,omitempty | maxHeaderBytes,omitempty | int | MaxHeaderBytes 对 \[http.Server.MaxHeaderBytes] 字段<br /> |
| recovery,omitempty | recovery,omitempty | recovery,attr,omitempty | recovery,omitempty | int | Recovery 拦截 panic 时反馈给客户端的状态码<br />NOTE: 这些设置对所有路径均有效，但会被 \[web.Routers.New] 的参数修改。<br /> |
| headers,omitempty | headers,omitempty | headers&gt;header,omitempty | headers,omitempty | [headerConfig](#headerconfig) | 自定义报头功能<br />报头会输出到包括 404 在内的所有请求返回。可以为空。<br />NOTE: 如果是与 CORS 相关的定义，则可能在 CORS 字段的定义中被修改。<br />NOTE: 报头内容可能会被后续的中间件修改。<br /> |
| cors,omitempty | cors,omitempty | cors,omitempty | cors,omitempty | [corsConfig](#corsconfig) | 自定义[跨域请求](https://developer.mozilla.org/zh-CN/docs/Web/HTTP/cors)设置项<br />NOTE: 这些设置对所有路径均有效，但会被 \[web.Routers.New] 的参数修改。<br /> |
| trace,omitempty | trace,omitempty | trace,omitempty | trace,omitempty | string | Trace 是否启用 TRACE 请求<br />可以有以下几种值：<br />  - disable 禁用 TRACE 请求；<br />  - body 启用 TRACE，且在返回内容中包含了请求端的 body 内容；<br />  - nobody 启用 TRACE，但是在返回内容中不包含请求端的 body 内容；<br />默认为 disable。<br />NOTE: 这些设置对所有路径均有效，但会被 \[web.Routers.New] 的参数修改。<br /> |



## certificateConfig




| JSON | YAML | XML | TOML | 类型 | 描述 |
|------|------|-----|------|------------------|------------------|
| cert,omitempty | cert,omitempty | cert,omitempty | cert,omitempty | string | 公钥文件地址<br /> |
| key,omitempty | key,omitempty | key,omitempty | key,omitempty | string | 私钥文件地址<br /> |



## acmeConfig




| JSON | YAML | XML | TOML | 类型 | 描述 |
|------|------|-----|------|------------------|------------------|
| domains | domains | domain | domain | string | 申请的域名列表<br /> |
| cache | cache | cache | cache | string | acme 缓存目录<br /> |
| email,omitempty | email,omitempty | email,omitempty | email,omitempty | string | 申请者邮箱<br /> |
| renewBefore,omitempty | renewBefore,omitempty | renewBefore,attr,omitempty | renewBefore,omitempty | uint | 定义提早几天开始续订，如果为 0 表示提早 30 天。<br /> |



## Duration

Duration 表示时间段<br />封装 [time.Duration](/time#Duration) 以实现对 JSON、XML、TOML 和 YAML 的解析<br />


## headerConfig




| JSON | YAML | XML | TOML | 类型 | 描述 |
|------|------|-----|------|------------------|------------------|
| key | key | key,attr | key | string | 报头名称<br /> |
| value | value | ,chardata | value | string | 报头对应的值<br /> |



## corsConfig




| JSON | YAML | XML | TOML | 类型 | 描述 |
|------|------|-----|------|------------------|------------------|
| origins,omitempty | origins,omitempty | origins&gt;origin,omitempty | origins,omitempty | string | 指定跨域中的 Access-Control-Allow-Origin 报头内容<br />如果为空，表示禁止跨域请示，如果包含了 \*，表示允许所有。<br /> |
| allowHeaders,omitempty | allowHeaders,omitempty | allowHeaders&gt;header,omitempty | allowHeaders,omitempty | string | 表示 Access-Control-Allow-Headers 报头内容<br /> |
| exposedHeaders,omitempty | exposedHeaders,omitempty | exposedHeaders&gt;header,omitempty | exposedHeaders,omitempty | string | 表示 Access-Control-Expose-Headers 报头内容<br /> |
| maxAge,omitempty | maxAge,omitempty | maxAge,attr,omitempty | maxAge,omitempty | int | 表示 Access-Control-Max-Age 报头内容<br />有以下几种取值：<br />  - 0 不输出该报头，默认值；<br />  - \-1 表示禁用；<br />  - 其它 >= -1 的值正常输出数值；<br /> |
| allowCredentials,omitempty | allowCredentials,omitempty | allowCredentials,attr,omitempty | allowCredentials,omitempty | bool | 表示 Access-Control-Allow-Credentials 报头内容<br /> |



## cacheConfig




| JSON | YAML | XML | TOML | 类型 | 描述 |
|------|------|-----|------|------------------|------------------|
| type | type | type,attr | type | string | 表示缓存的方式<br />该值可通过 \[RegisterCache] 注册， 默认支持以下几种：<br />  - memory 以内存作为缓存；<br />  - memcached 以 memcached 作为缓存；<br />  - redis 以 redis 作为缓存；<br /> |
| dsn | dsn | dsn | dsn | string | 表示连接缓存服务器的参数<br />不同类型其参数是不同的，以下是对应的格式说明：<br />  - memory: 不需要参数；<br />  - memcached: 则为服务器列表，多个服务器，以分号作为分隔；<br />  - redis: 符合 [Redis URI scheme](https://www.iana.org/assignments/uri-schemes/prov/redis) 的字符串；<br /> |



## compressConfig




| JSON | YAML | XML | TOML | 类型 | 描述 |
|------|------|-----|------|------------------|------------------|
| types | types | type | types | string | Type content-type 的值<br />可以带通配符，比如 text/\* 表示所有 text/ 开头的 content-type 都采用此压缩方法。<br /> |
| id | id | id,attr | id | string | IDs 压缩方法的 ID 列表<br />这些 ID 值必须是由 \[RegisterCompress] 注册的，否则无效，默认情况下支持以下类型：<br />  - deflate-default<br />  - deflate-best-compression<br />  - deflate-best-speed<br />  - gzip-default<br />  - gzip-best-compression<br />  - gzip-best-speed<br />  - compress-lsb-8<br />  - compress-msb-8<br />  - br-default<br />  - br-best-compression<br />  - br-best-speed<br />  - zstd-default<br /> |



## mimetypeConfig




| JSON | YAML | XML | TOML | 类型 | 描述 |
|------|------|-----|------|------------------|------------------|
| type | type | type,attr | type | string | 编码名称<br />比如 application/xml 等<br /> |
| problem,omitempty | problem,omitempty | problem,attr,omitempty | problem,omitempty | string | 返回错误代码是的 mimetype<br />比如正常情况下如果是 application/json，那么此值可以是 application/problem+json。 如果为空，表示与 Type 相同。<br /> |
| target | target | target,attr | target | string | 实际采用的解码方法<br />由 \[RegisterMimetype] 注册而来。默认可用为：<br />  - xml<br />  - cbor<br />  - json<br />  - form<br />  - html<br />  - gob<br />  - yaml<br />  - nop  没有具体实现的方法，对于上传等需要自行处理的情况可以指定此值。<br /> |



## registryConfig

registryConfig 注册服务中心的配置项<br />


| JSON | YAML | XML | TOML | 类型 | 描述 |
|------|------|-----|------|------------------|------------------|
| type | type | type | type | string | 配置的保存类型<br />该类型可通过 \[RegisterRegistryType] 进行注册，默认情况下支持以下类型：<br />  - cache 以缓存系统作为储存类型；<br /> |
| strategy | strategy | strategy | strategy | string | 负载均衡的方案<br />可通过 \[RegisterStrategy] 进行注册，默认情况下支持以下类型：<br />  - random 随机；<br />  - weighted-random 带权重的随机；<br />  - round-robin 轮循；<br />  - weighted-round-robin 带权重的轮循；<br /> |
| args,omitempty | args,omitempty | args&gt;arg,omitempty | args,omitempty | string | 传递 Type 的额外参数<br />会根据 args 的不同而不同：<br />  - cache 仅支持一个参数，为 [time.ParseDuration](/time#ParseDuration) 可解析的字符串；<br /> |



## mapperConfig




| JSON | YAML | XML | TOML | 类型 | 描述 |
|------|------|-----|------|------------------|------------------|
| name | name | name | name | string | 微服务名称<br /> |
| matcher | matcher | matcher | matcher | string | 判断某个请求是否进入当前微服务的方法<br />该值可通过 \[RegisterRouterMatcher] 注册，默认情况下支持以下类型：<br />  - hosts 只限定域名；<br />  - prefix 包含特定前缀的访问地址；<br />  - version 在 accept 中指定的特定的版本号才行；<br />  - any 任意；<br /> |
| args,omitempty | args,omitempty | args&gt;arg,omitempty | args,omitempty | string | 传递 Matcher 的额外参数<br />会根据 Matcher 的不同而不同：<br />  - hosts 以逗号分隔的域名列表；<br />  - prefix 以逗号分隔的 URL 前缀列表；<br />  - version 允许放行的版本号列表(以逗号分隔)，这些版本号出现在 accept 报头；<br />  - any 不需要参数；<br /> |



## T

未找到类型的相关文档


