# RaceMate Desktop

A desktop client application that reads simracing telemetry data from Assetto Corsa Competizione (ACC) and provides user authentication via Firebase.

## Features

- ACC telemetry data collection
- User authentication with Firebase
- Persistent login sessions
- System tray integration

## Prerequisites

- Go 1.16 or higher
- Make
- Windows operating system (currently only Windows is supported)
- Firebase project (for authentication)

## Environment Setup

The application requires Firebase configuration. Create a `.env` file in the root directory based on the provided `.env.example` file:

```
FIREBASE_API_KEY=your_api_key
FIREBASE_AUTH_DOMAIN=your_project.firebaseapp.com
FIREBASE_PROJECT_ID=your_project_id
FIREBASE_STORAGE_BUCKET=your_project.appspot.com
FIREBASE_MESSAGING_SENDER_ID=your_sender_id
FIREBASE_APP_ID=your_app_id
FIREBASE_MEASUREMENT_ID=your_measurement_id
```

## Building and Running

The project includes a Makefile with several useful tasks:

### Build the Application

```bash
# Production build
make build

# Development build
make build-dev
```

Built binaries will be placed in the `build` directory.

### Run the Application

```bash
# Run production build
make run

# Run development build
make run-dev
```

### Clean Build Artifacts

```bash
make clean
```

### Package for Windows

Create a distributable Windows package using Fyne:

```bash
make package-windows
```

## Development

When developing, use the `make run-dev` command which builds the application with the development configuration and runs it immediately.

## Data Storage

The application stores data in the following locations:

- Application data: `%AppData%\RaceMate`
- Telemetry data: `%AppData%\RaceMate\upload`
- Processed data: `%AppData%\RaceMate\uploaded`
- Log files: `%AppData%\RaceMate\logs`
- Authentication data: `%AppData%\RaceMate\auth`

## License

TBD
