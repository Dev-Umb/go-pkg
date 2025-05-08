package errno

var (
	// OK 表示成功
	OK = &Errno{Code: 0, Message: "成功"}

	// 系统级错误 (10000-10999)
	InternalServerError = &Errno{Code: 10001, Message: "内部服务器错误"}

	ConvertError  = &Errno{Code: 10003, Message: "数据转换错误"}
	NotFoundError = &Errno{Code: 10004, Message: "资源未找到"}
	TimeoutError  = &Errno{Code: 10005, Message: "请求超时"}
	ErrBind       = &Errno{Code: 10006, Message: "请求参数绑定失败"}

	// 数据库错误 (11000-11999)
	ErrDatabase            = &Errno{Code: 11000, Message: "数据库操作错误"}
	ErrDatabaseConnect     = &Errno{Code: 11001, Message: "数据库连接失败"}
	ErrDatabaseQuery       = &Errno{Code: 11002, Message: "数据库查询错误"}
	ErrDatabaseInsert      = &Errno{Code: 11003, Message: "数据库插入错误"}
	ErrDatabaseUpdate      = &Errno{Code: 11004, Message: "数据库更新错误"}
	ErrDatabaseDelete      = &Errno{Code: 11005, Message: "数据库删除错误"}
	ErrDatabaseTransaction = &Errno{Code: 11006, Message: "数据库事务操作失败"}

	// Redis缓存错误 (12000-12999)
	ErrRedis        = &Errno{Code: 12000, Message: "Redis操作错误"}
	ErrRedisConnect = &Errno{Code: 12001, Message: "Redis连接失败"}
	ErrRedisSet     = &Errno{Code: 12002, Message: "Redis设置值失败"}
	ErrRedisGet     = &Errno{Code: 12003, Message: "Redis获取值失败"}
	ErrRedisDelete  = &Errno{Code: 12004, Message: "Redis删除值失败"}
	ErrRedisExpire  = &Errno{Code: 12005, Message: "Redis设置过期时间失败"}

	// 认证与授权错误 (13000-13999)
	ErrUnauthorizedError  = &Errno{Code: 13001, Message: "未授权，请登录"}
	InvalidTokenError     = &Errno{Code: 13002, Message: "无效的令牌"}
	ExpiredTokenError     = &Errno{Code: 13003, Message: "令牌已过期"}
	GenerateJwtTokenError = &Errno{Code: 13004, Message: "生成JWT令牌错误"}
	TokenValidationError  = &Errno{Code: 13005, Message: "令牌验证失败"}
	PermissionDeniedError = &Errno{Code: 13006, Message: "权限不足，拒绝访问"}

	// 参数验证错误 (14000-14999)
	InvalidParamsError   = &Errno{Code: 14001, Message: "无效的参数"}
	MissingParamError    = &Errno{Code: 14002, Message: "缺少必要参数"}
	InvalidFormatError   = &Errno{Code: 14003, Message: "参数格式不正确"}
	ValueOutOfRangeError = &Errno{Code: 14004, Message: "参数值超出范围"}
	BindRequestError     = &Errno{Code: 14005, Message: "请求参数绑定失败"}

	// 第三方服务错误 (15000-15999)
	ThirdPartyServiceError = &Errno{Code: 15001, Message: "第三方服务错误"}
	APIGatewayError        = &Errno{Code: 15002, Message: "API网关错误"}
	ExternalAPIError       = &Errno{Code: 15003, Message: "外部API调用失败"}

	// 限流与并发控制错误 (16000-16999)
	RateLimitError        = &Errno{Code: 16001, Message: "请求频率超过限制"}
	ConcurrencyLimitError = &Errno{Code: 16002, Message: "并发请求数超过限制"}
	ServerBusyError       = &Errno{Code: 16003, Message: "服务繁忙，请稍后再试"}

	// 文件与上传错误 (17000-17999)
	FileUploadError    = &Errno{Code: 17001, Message: "文件上传失败"}
	FileDownloadError  = &Errno{Code: 17002, Message: "文件下载失败"}
	FileFormatError    = &Errno{Code: 17003, Message: "文件格式不支持"}
	FileSizeLimitError = &Errno{Code: 17004, Message: "文件大小超过限制"}
	FileSaveError      = &Errno{Code: 17005, Message: "文件保存失败"}

	// 验证码错误 (18000-18999)
	AuthCodeError         = &Errno{Code: 18001, Message: "验证码错误"}
	AuthCodeExpiredError  = &Errno{Code: 18002, Message: "验证码已过期"}
	AuthCodeGenerateError = &Errno{Code: 18003, Message: "验证码生成失败"}
	AuthCodeSendError     = &Errno{Code: 18004, Message: "验证码发送失败"}

	// 密码相关错误 (19000-19999)
	PasswordError         = &Errno{Code: 19001, Message: "密码错误"}
	PasswordFormatError   = &Errno{Code: 19002, Message: "密码格式不符合要求"}
	PasswordResetError    = &Errno{Code: 19003, Message: "密码重置失败"}
	PasswordNotMatchError = &Errno{Code: 19004, Message: "两次密码不匹配"}

	// SendEmailError 邮件相关错误
	SendEmailError = &Errno{Code: 19500, Message: "邮件发送失败！"}
)

// 业务错误码 (20000-99999)

// 用户相关错误码 (20000-20999)
var (
	ErrUserNotFound      = &Errno{Code: 20001, Message: "用户不存在"}
	ErrUserAlreadyExist  = &Errno{Code: 20002, Message: "用户已存在"}
	ErrUserCreateFailed  = &Errno{Code: 20003, Message: "创建用户失败"}
	ErrUserUpdateFailed  = &Errno{Code: 20004, Message: "更新用户信息失败"}
	ErrUserDeleteFailed  = &Errno{Code: 20005, Message: "删除用户失败"}
	ErrUserIDInvalid     = &Errno{Code: 20006, Message: "用户ID无效"}
	ErrUserLocked        = &Errno{Code: 20007, Message: "用户账号已锁定"}
	ErrUserDisabled      = &Errno{Code: 20008, Message: "用户账号已禁用"}
	ErrUserPhoneInvalid  = &Errno{Code: 20009, Message: "手机号格式不正确"}
	ErrUserEmailInvalid  = &Errno{Code: 20010, Message: "邮箱格式不正确"}
	ErrUserAvatarInvalid = &Errno{Code: 20011, Message: "头像格式不支持"}
	ErrUserLoginFailed   = &Errno{Code: 20012, Message: "用户登录失败"}
	ErrUserLogoutFailed  = &Errno{Code: 20013, Message: "用户登出失败"}
)

// 令牌相关错误码 (21000-21999)
var (
	ErrTokenInvalid      = &Errno{Code: 21001, Message: "无效的令牌"}
	ErrTokenExpired      = &Errno{Code: 21002, Message: "令牌已过期"}
	ErrTokenRevoked      = &Errno{Code: 21003, Message: "令牌已被撤销"}
	ErrTokenMalformed    = &Errno{Code: 21004, Message: "令牌格式错误"}
	ErrTokenMissing      = &Errno{Code: 21005, Message: "缺少令牌"}
	ErrTokenGenerate     = &Errno{Code: 21006, Message: "令牌生成失败"}
	ErrTokenValidate     = &Errno{Code: 21007, Message: "令牌验证失败"}
	ErrTokenUserMismatch = &Errno{Code: 21008, Message: "令牌与用户不匹配"}
)

// 网络与RPC错误码 (30000-30999)
var (
	ErrRPCConnection         = &Errno{Code: 30001, Message: "RPC连接失败"}
	ErrRPCTimeout            = &Errno{Code: 30002, Message: "RPC调用超时"}
	ErrRPCInvalidResponse    = &Errno{Code: 30003, Message: "RPC响应无效"}
	ErrRPCServiceUnavailable = &Errno{Code: 30004, Message: "RPC服务不可用"}
	ErrNetworkUnavailable    = &Errno{Code: 30005, Message: "网络不可用"}
	ErrNetworkTimeout        = &Errno{Code: 30006, Message: "网络请求超时"}
	ErrDNSResolution         = &Errno{Code: 30007, Message: "DNS解析失败"}
)

// 配置错误码 (40000-40999)
var (
	ErrConfigNotFound = &Errno{Code: 40001, Message: "配置未找到"}
	ErrConfigInvalid  = &Errno{Code: 40002, Message: "配置无效"}
	ErrConfigParse    = &Errno{Code: 40003, Message: "配置解析失败"}
	ErrConfigLoad     = &Errno{Code: 40004, Message: "配置加载失败"}
)

// 错误映射表，用于将不同类型的error映射到适当的errno
var ErrorMap = map[string]*Errno{
	// 数据库错误映射
	"record not found":            ErrUserNotFound,
	"duplicate key":               ErrUserAlreadyExist,
	"Error 1062: Duplicate entry": ErrUserAlreadyExist,
	"database is closed":          ErrDatabaseConnect,
	"connection refused":          ErrDatabaseConnect,

	// token错误映射
	"token is expired":                             ErrTokenExpired,
	"signature is invalid":                         ErrTokenInvalid,
	"token contains an invalid number of segments": ErrTokenMalformed,

	// Redis错误映射
	"redis: connection pool timeout": ErrRedisConnect,
	"redis: connection closed":       ErrRedisConnect,
	"redis: nil":                     ErrRedisGet,

	// 网络错误映射
	"context deadline exceeded": TimeoutError,
	"no such host":              ErrNetworkUnavailable,
	"connection reset by peer":  ErrNetworkUnavailable,
}
