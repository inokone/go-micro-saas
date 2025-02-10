# Prerequisites

Before running the application, you need to set up the following third-party services and obtain their respective API keys/credentials:

## Database

- PostgreSQL database server
  - Required for storing user data, authentication information, and history
  - Configuration variables needed:
    - `DB_HOST`: Database host
    - `DB_PORT`: Database port
    - `DB_NAME`: Database name
    - `DB_USER`: Database username
    - `DB_PASS`: Database password
    - `DB_SSL_MODE`: SSL mode (default: "disable")
    - `DB_SSL_CERT`: SSL certificate path (if SSL is enabled)

## Authentication Services

### Google Authentication

1. Create a project in the [Google Cloud Console](https://console.cloud.google.com/)
2. Enable Google OAuth 2.0
3. Create OAuth 2.0 credentials
4. Configure the following variables:
   - `GOOGLE_AUTH_KEY`: Google OAuth client ID
   - `GOOGLE_AUTH_SECRET`: Google OAuth client secret

### Facebook Authentication

1. Create an app in the [Facebook Developers Console](https://developers.facebook.com/)
2. Enable Facebook Login
3. Configure the following variables:
   - `FACEBOOK_AUTH_KEY`: Facebook OAuth client ID
   - `FACEBOOK_AUTH_SECRET`: Facebook OAuth client secret

### Google reCAPTCHA Enterprise

1. Enable reCAPTCHA Enterprise in your Google Cloud project
2. Create a reCAPTCHA Enterprise key
3. Configure the following variables:
   - `GOOGLE_APPLICATION_CREDENTIALS`: Path to Google Cloud service account credentials file
   - `GOOGLE_PROJECT_ID`: Google Cloud project ID
   - `GOOGLE_RECAPTCHA_KEY`: reCAPTCHA site key

## Email Service

SMTP server configuration for sending emails (e.g., confirmation emails, password reset):

- `MAIL_SMTP_ADDRESS`: SMTP server address
- `MAIL_SMTP_PORT`: SMTP server port
- `MAIL_SMTP_USER`: SMTP username
- `MAIL_SMTP_PASSWORD`: SMTP password
- `MAIL_NO_REPLY_ADDRESS`: No-reply email address for sending system emails
- `APPLICATION_NAME`: Application name to use in email templates

## Analytics (Optional)

- Statsig analytics integration:
  - `STATSIG_SERVER_SECRET_KEY`: Statsig server secret key

## Security Configuration

JWT (JSON Web Token) configuration:

- `JWT_SIGN_SECRET`: Secret key for signing JWT tokens
- `JWT_EXPIRATION_HOURS`: JWT token expiration time in hours (default: 24)
- `JWT_COOKIE_SECURE`: Whether to use secure cookies (default: true)

## TLS Configuration (Optional)

For HTTPS support:

- `TLS_CERT_PATH`: Path to TLS certificate
- `TLS_KEY_PATH`: Path to TLS private key

## Application URLs

- `FRONTEND_ROOT`: URL of the frontend application (e.g., "https://example.com")
- `BACKEND_ROOT`: URL of the backend API (e.g., "https://api.example.com")

## Environment Setup

1. Create an `app.env` file in one of the following locations:
   - Current directory
   - `/etc/microsaas/`
   - `$HOME/.microsaas`
2. Add all the required configuration variables to the file
3. Ensure all necessary credentials files are in place and properly referenced in the configuration

## Minimum Required Configuration

At a minimum, you need to configure:

1. Database connection
2. Email service
3. Google reCAPTCHA
4. JWT signing secret
5. Frontend and backend URLs

The other services (Google/Facebook auth, TLS, analytics) can be configured based on your application's needs. 
