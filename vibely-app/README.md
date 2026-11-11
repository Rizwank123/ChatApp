# Vibely

Expo SDK 54 client for the ChatApp social backend. It targets web, Android, and iOS.

## Run

```bash
npm install
npm run web
```

Set API URLs when the backend is not on localhost:

```bash
EXPO_PUBLIC_API_URL=http://localhost:7700/api/v1 EXPO_PUBLIC_WS_URL=ws://localhost:7700/ws npm start
```

Android emulators usually need `http://10.0.2.2:7700/api/v1`.
