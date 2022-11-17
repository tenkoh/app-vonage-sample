# app-vonage-sample
sample web application using Vonage for 2FA

## Overview

![overview](https://camo.qiitausercontent.com/f8cc8a832cb1931a63b8f6b00719311c53ff6a5e/68747470733a2f2f71696974612d696d6167652d73746f72652e73332e61702d6e6f727468656173742d312e616d617a6f6e6177732e636f6d2f302f313535333139312f35613262316239392d663537362d396431362d646363362d3661656363393832313363352e706e67)

1. When you access to the top page at the first time, you are a non-verified user.
2. Enter your phone number and submit. Then you will get a SMS giving 4 digits PIN code.
3. Enter the PIN code and submit. If succeed, you are redirected to the top page at a verified user.

## Usage
1. Set environment variables. Vonage API key & secret are provided on Vonage. (need to register)
   - VONAGE_API_KEY
   - VONAGE_API_SECRET
   - SESSION_SECRET (optional)
2. Clone this repository.
3. `go run .`
4. Access to `http://localhost:1323`.

## License
MIT

## Author
tenkoh
