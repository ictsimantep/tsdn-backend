package templates

const ResetEmailTemplateStd = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Password Reset</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            background-color: #f4f4f7;
            color: #51545e;
            margin: 0;
            padding: 0;
        }
        .email-container {
            width: 100%;
            background-color: #f4f4f7;
            padding: 20px;
        }
        .email-content {
            max-width: 600px;
            margin: 0 auto;
            background-color: #ffffff;
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
            padding: 40px;
        }
        .email-header {
            text-align: center;
            padding-bottom: 20px;
        }
        .email-header img {
            width: 100px;
        }
        .email-body {
            text-align: center;
            padding: 0 20px;
        }
        .email-body h1 {
            color: #333333;
            font-size: 24px;
            font-weight: bold;
            margin-bottom: 20px;
        }
        .email-body p {
            font-size: 16px;
            line-height: 1.6;
            margin-bottom: 30px;
            color: #51545e;
        }
        .email-button {
            text-align: center;
            margin-bottom: 30px;
        }
        .email-button a {
            background-color: #007bff;
            color: #ffffff;
            padding: 12px 24px;
            text-decoration: none;
            border-radius: 5px;
            font-size: 16px;
        }
        .email-footer {
            text-align: center;
            font-size: 12px;
            color: #999999;
            margin-top: 40px;
        }
        .email-footer a {
            color: #007bff;
            text-decoration: none;
        }
        .email-footer p {
            margin-top: 0;
        }
        @media only screen and (max-width: 600px) {
            .email-content {
                padding: 20px;
            }
            .email-button a {
                font-size: 14px;
            }
        }
    </style>
</head>
<body>
    <div class="email-container">
        <div class="email-content">
            <div class="email-header">
                <img src="https://example.com/logo.png" alt="Company Logo">
            </div>
            <div class="email-body">
                <h1>Password Reset Request</h1>
                <p>
                    Hello, <br>
                    You are receiving this email because we received a password reset request for your account. Please click the button below to reset your password. If you did not request a password reset, you can safely ignore this email.
                </p>
                <div class="email-button">
                    <a href="{{RESET_LINK}}" target="_blank">Reset Password</a>
                </div>
                <p>
                    This password reset link will expire in 1 hour. <br>
                    If you have any issues, feel free to contact us.
                </p>
            </div>
            <div class="email-footer">
                <p>&copy; 2024 Company Name. All rights reserved.</p>
                <p>
                    <a href="https://example.com">Visit our website</a> |
                    <a href="https://example.com/contact">Contact us</a>
                </p>
            </div>
        </div>
    </div>
</body>
</html>
`
