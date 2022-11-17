# app-vonage-sample
sample web application using Vonage for 2FA

## Overview

![image.png](https://qiita-image-store.s3.ap-northeast-1.amazonaws.com/0/1553191/299ccce4-1e04-001b-9882-d1331b57d5b6.png)

1. When you access to the top page at the first time, you are a non-verified user.
2. Enter your phone number and submit. Then you will get a SMS giving 4 digits PIN code.
3. Enter the PIN code and submit. If succeed, you are redirected to the top page as a verified user.

**NOTICE** : You have to pay fee for each verification with the Vonage service. (And perhaps you would need to pay some reception payment for SMS)

## Usage
1. Set environment variables. Vonage API key & secret are provided on Vonage. (need to register)
   - VONAGE_API_KEY
   - VONAGE_API_SECRET
   - SESSION_AUTH_KEY (optional)
2. Clone this repository.
3. `go run .`
4. Access to `http://localhost:1323`.

## License
MIT

## Author
tenkoh
