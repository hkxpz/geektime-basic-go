.PHNOY:signup
signup:
	@wrk -t1 -d1s -c2 -s ./signup.lua http://localhost:8081/users/signup

.PHNOY:login
login:
	@wrk -t1 -d1s -c2 -s ./login.lua http://localhost:8081/users/login

.PHNOY:profile
profile:
	@wrk -t1 -d1s -c2 -s ./profile.lua http://localhost:8081/users/profile