.PHNOY:signup
signup:
	@wrk -t1 -d1s -c1 -s ./signup.lua http://localhost:8080/users/signup

.PHNOY:login
login:
	@wrk -t1 -d1m -c10 -s ./login.lua http://localhost:8080/users/login

.PHNOY:profile
profile:
	@wrk -t1 -d1s -c2 -s ./profile.lua http://localhost:8080/users/profile

.PHNOY:article.like
article.like:
	@wrk -t1 -d10 -c1 -s ./article/like.lua http://localhost:8080/articles/pub/like