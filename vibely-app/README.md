# Vibely

Expo SDK 54 client for the ChatApp social backend. It targets web, Android, and iOS.

## Run

```bash
npm install
npm run web
```

The root Docker Compose stack also builds and serves the web frontend at `http://localhost:8081`. Set `FRONTEND_API_URL` and `FRONTEND_WS_URL` in the root `.env` before rebuilding if the browser should reach the API through a different host or port. Local Docker defaults allow HTTP and WS; production builds should use HTTPS and WSS and set `FRONTEND_ALLOW_INSECURE_API=false`.

Set API URLs when the backend is not on localhost:

```bash
EXPO_PUBLIC_API_URL=http://localhost:7700/api/v1 EXPO_PUBLIC_WS_URL=ws://localhost:7700/ws npm start
```

Android emulators usually need `http://10.0.2.2:7700/api/v1`.
