import AsyncStorage from "@react-native-async-storage/async-storage";
import React, {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
} from "react";
import { Platform } from "react-native";

import { readJwtClaims } from "@/api/jwt";
import type { AppUser, Message, Post } from "@/api/types";

type Toast = {
  message: string;
  kind?: "info" | "error" | "success";
};

type AppContextValue = {
  session?: AppUser;
  setSession: (session?: AppUser) => void;
  isRestoringSession: boolean;
  posts: Post[];
  setPosts: (posts: Post[]) => void;
  upsertPost: (post: Post) => void;
  messages: Record<string, Message[]>;
  appendMessage: (message: Message) => void;
  toast?: Toast;
  showToast: (toast: Toast) => void;
  clearToast: () => void;
};

const SESSION_STORAGE_KEY = "@vibely_session";

function isSessionValid(session?: AppUser | null): session is AppUser {
  if (!session || !session.token) {
    return false;
  }
  try {
    const claims = readJwtClaims(session.token);
    if (claims.exp && claims.exp * 1000 <= Date.now()) {
      return false;
    }
    return true;
  } catch {
    return false;
  }
}

function getStoredSessionSync(): AppUser | undefined {
  if (typeof window !== "undefined" && window.localStorage) {
    try {
      const raw = window.localStorage.getItem(SESSION_STORAGE_KEY);
      if (raw) {
        const parsed = JSON.parse(raw) as AppUser;
        if (isSessionValid(parsed)) {
          return parsed;
        }
        window.localStorage.removeItem(SESSION_STORAGE_KEY);
      }
    } catch {
      // ignore JSON or localStorage errors
    }
  }
  return undefined;
}

const AppContext = createContext<AppContextValue | undefined>(undefined);

export function AppProvider({ children }: { children: React.ReactNode }) {
  const initialSession = useMemo(() => getStoredSessionSync(), []);
  const [session, setSessionState] = useState<AppUser | undefined>(initialSession);
  const [isRestoringSession, setIsRestoringSession] = useState<boolean>(
    () => !initialSession && Platform.OS !== "web",
  );
  const [posts, setPosts] = useState<Post[]>([]);
  const [messages, setMessages] = useState<Record<string, Message[]>>({});
  const [toast, setToast] = useState<Toast>();

  useEffect(() => {
    let isMounted = true;
    (async () => {
      try {
        const raw = await AsyncStorage.getItem(SESSION_STORAGE_KEY);
        if (raw && isMounted) {
          const parsed = JSON.parse(raw) as AppUser;
          if (isSessionValid(parsed)) {
            setSessionState(parsed);
          } else {
            await AsyncStorage.removeItem(SESSION_STORAGE_KEY);
            if (typeof window !== "undefined" && window.localStorage) {
              window.localStorage.removeItem(SESSION_STORAGE_KEY);
            }
            setSessionState(undefined);
          }
        }
      } catch {
        // ignore
      } finally {
        if (isMounted) {
          setIsRestoringSession(false);
        }
      }
    })();

    return () => {
      isMounted = false;
    };
  }, []);

  const setSession = useCallback((nextSession?: AppUser) => {
    setSessionState(nextSession);
    if (nextSession) {
      AsyncStorage.setItem(SESSION_STORAGE_KEY, JSON.stringify(nextSession)).catch(() => {});
      if (typeof window !== "undefined" && window.localStorage) {
        try {
          window.localStorage.setItem(SESSION_STORAGE_KEY, JSON.stringify(nextSession));
        } catch {
          // ignore
        }
      }
    } else {
      AsyncStorage.removeItem(SESSION_STORAGE_KEY).catch(() => {});
      if (typeof window !== "undefined" && window.localStorage) {
        try {
          window.localStorage.removeItem(SESSION_STORAGE_KEY);
        } catch {
          // ignore
        }
      }
    }
  }, []);

  const upsertPost = useCallback((post: Post) => {
    setPosts((current) => {
      const exists = current.some((item) => item.id === post.id);
      if (exists) {
        return current.map((item) => (item.id === post.id ? post : item));
      }
      return [post, ...current];
    });
  }, []);

  const appendMessage = useCallback((message: Message) => {
    setMessages((current) => ({
      ...current,
      [message.conversation_id]: [
        ...(current[message.conversation_id] ?? []),
        message,
      ],
    }));
  }, []);

  const showToast = useCallback((nextToast: Toast) => {
    setToast(nextToast);
  }, []);

  const clearToast = useCallback(() => {
    setToast(undefined);
  }, []);

  const value = useMemo<AppContextValue>(
    () => ({
      session,
      setSession,
      isRestoringSession,
      posts,
      setPosts,
      upsertPost,
      messages,
      appendMessage,
      toast,
      showToast,
      clearToast,
    }),
    [
      appendMessage,
      clearToast,
      isRestoringSession,
      messages,
      posts,
      session,
      setSession,
      showToast,
      toast,
      upsertPost,
    ],
  );

  return <AppContext.Provider value={value}>{children}</AppContext.Provider>;
}

export function useApp() {
  const value = useContext(AppContext);
  if (!value) {
    throw new Error("useApp must be used inside AppProvider");
  }
  return value;
}
