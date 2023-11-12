package errs

// User 部分，模块代码使用 01
const (
	// UserInvalidInput 这是一个非常含糊的错误码，代表用户相关的API参数不对
	UserInvalidInput = 401001
	// UserInternalServerError 这是一个非常函数的错误码。代表用户模块系统内部错误
	UserInternalServerError = 501001
	// 假设说我们需要进一步关心别的错误

	// UserInvalidOrPassword 用户输入的账号或者密码不对
	UserInvalidOrPassword = 401002
	// UserDuplicateEmail 邮箱冲突
	UserDuplicateEmail = 401003
)

// Article 部分，模块代码使用 02
const (
	// ArticleInvalidInput 含糊的输入错误
	ArticleInvalidInput        = 402001
	ArticleInternalServerError = 502001
)
