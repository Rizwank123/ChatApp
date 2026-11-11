import React, { useEffect, useRef, useState } from "react";
import {
  ActivityIndicator,
  KeyboardAvoidingView,
  Platform,
  Pressable,
  ScrollView,
  StyleSheet,
  Text,
  TextInput,
  View,
} from "react-native";

import { api, apiConfig } from "../api/client";
import { readJwtClaims } from "../api/jwt";
import { useApp } from "../state/AppContext";
import { colors } from "../theme/colors";

export function AuthScreen() {
  const { setSession, showToast } = useApp();
  const [mode, setMode] = useState<"login" | "register">("login");
  const [firstName, setFirstName] = useState("");
  const [lastName, setLastName] = useState("");
  const [userName, setUserName] = useState("");
  const [password, setPassword] = useState("");
  const [busy, setBusy] = useState(false);

  // Username validation & availability state
  const [checkingUsername, setCheckingUsername] = useState(false);
  const [usernameStatus, setUsernameStatus] = useState<{
    checked: boolean;
    available: boolean;
    reason?: string;
  } | null>(null);
  const debounceTimer = useRef<NodeJS.Timeout | null>(null);

  useEffect(() => {
    if (mode !== "register" || !userName.trim()) {
      setUsernameStatus(null);
      setCheckingUsername(false);
      return;
    }

    const clean = userName.trim().toLowerCase();
    if (clean.length < 3) {
      setUsernameStatus({
        checked: true,
        available: false,
        reason: "Username must be at least 3 characters",
      });
      return;
    }

    if (debounceTimer.current) {
      clearTimeout(debounceTimer.current);
    }

    setCheckingUsername(true);
    debounceTimer.current = setTimeout(async () => {
      try {
        const res = await api.checkUsername(clean);
        setUsernameStatus({
          checked: true,
          available: res.data.is_available,
          reason: res.data.reason,
        });
      } catch (err: any) {
        setUsernameStatus({
          checked: true,
          available: false,
          reason: err?.message || "Error validating username",
        });
      } finally {
        setCheckingUsername(false);
      }
    }, 400);

    return () => {
      if (debounceTimer.current) {
        clearTimeout(debounceTimer.current);
      }
    };
  }, [userName, mode]);

  async function submit() {
    setBusy(true);
    try {
      const cleanUsername = userName.trim().toLowerCase();
      if (mode === "register") {
        if (usernameStatus && !usernameStatus.available) {
          showToast({
            message: usernameStatus.reason || "Username is unavailable",
            kind: "error",
          });
          setBusy(false);
          return;
        }
        await api.register({
          first_name: firstName,
          last_name: lastName,
          user_name: cleanUsername,
          role: "USER",
          password,
        });
        showToast({ message: "Account created. Signing you in...", kind: "success" });
      }

      const result = await api.login({ username: cleanUsername, password });
      const claims = readJwtClaims(result.data.token);
      setSession({
        token: result.data.token,
        expiresIn: result.data.expires_in,
        userName: cleanUsername,
        personnelId: claims.personnel_id,
      });
    } catch (error) {
      showToast({
        message: error instanceof Error ? error.message : "Authentication failed",
        kind: "error",
      });
    } finally {
      setBusy(false);
    }
  }

  const isSubmitDisabled =
    busy ||
    !userName.trim() ||
    !password ||
    (mode === "register" &&
      (checkingUsername || (usernameStatus !== null && !usernameStatus.available)));

  return (
    <KeyboardAvoidingView
      behavior={Platform.OS === "ios" ? "padding" : undefined}
      style={styles.screen}
    >
      <ScrollView contentContainerStyle={styles.content}>
        {/* Instagram Brand Title */}
        <View style={styles.brandBlock}>
          <Text style={styles.logo}>Vibely</Text>
          <Text style={styles.tagline}>
            {mode === "login"
              ? "Sign in to see photos and videos from your friends."
              : "Sign up to see photos and videos from your friends."}
          </Text>
        </View>

        {/* Input Fields (Instagram style) */}
        <View style={styles.form}>
          {mode === "register" && (
            <View style={styles.row}>
              <View style={styles.half}>
                <TextInput
                  style={styles.input}
                  placeholder="First name"
                  placeholderTextColor={colors.textDim}
                  value={firstName}
                  onChangeText={setFirstName}
                />
              </View>
              <View style={styles.half}>
                <TextInput
                  style={styles.input}
                  placeholder="Last name"
                  placeholderTextColor={colors.textDim}
                  value={lastName}
                  onChangeText={setLastName}
                />
              </View>
            </View>
          )}

          <View>
            <TextInput
              style={styles.input}
              placeholder="Username, phone, or email"
              placeholderTextColor={colors.textDim}
              value={userName}
              onChangeText={(t) => setUserName(t.toLowerCase())}
              autoCapitalize="none"
              autoCorrect={false}
            />
            {mode === "register" && (
              <View style={styles.availabilityRow}>
                {checkingUsername ? (
                  <View style={styles.statusIndicator}>
                    <ActivityIndicator size="small" color={colors.primary} />
                    <Text style={styles.checkingText}>Checking availability...</Text>
                  </View>
                ) : usernameStatus?.checked ? (
                  <Text
                    style={[
                      styles.statusText,
                      usernameStatus.available ? styles.statusAvailable : styles.statusUnavailable,
                    ]}
                  >
                    {usernameStatus.available
                      ? "✓ Username is available"
                      : `✗ ${usernameStatus.reason || "Username is unavailable"}`}
                  </Text>
                ) : null}
              </View>
            )}
          </View>

          <TextInput
            style={styles.input}
            placeholder="Password"
            placeholderTextColor={colors.textDim}
            secureTextEntry
            value={password}
            onChangeText={setPassword}
          />

          {/* Instagram Blue Action Button */}
          <Pressable
            onPress={submit}
            disabled={isSubmitDisabled}
            style={[styles.primaryButton, isSubmitDisabled && styles.disabledButton]}
          >
            {busy ? (
              <ActivityIndicator size="small" color="#ffffff" />
            ) : (
              <Text style={styles.primaryButtonText}>
                {mode === "login" ? "Log in" : "Sign up"}
              </Text>
            )}
          </Pressable>
        </View>

        {/* OR Divider */}
        <View style={styles.orRow}>
          <View style={styles.orLine} />
          <Text style={styles.orText}>OR</Text>
          <View style={styles.orLine} />
        </View>

        {/* Switch Mode Footer */}
        <View style={styles.footerWrap}>
          <Text style={styles.footerPrompt}>
            {mode === "login" ? "Don't have an account? " : "Have an account? "}
          </Text>
          <Pressable
            onPress={() => setMode((m) => (m === "login" ? "register" : "login"))}
            hitSlop={8}
          >
            <Text style={styles.footerAction}>
              {mode === "login" ? "Sign up" : "Log in"}
            </Text>
          </Pressable>
        </View>

        {/* From Meta / From Vibely footer */}
        <View style={styles.metaBranding}>
          <Text style={styles.fromText}>from</Text>
          <Text style={styles.brandFooter}>VIBELY</Text>
        </View>
      </ScrollView>
    </KeyboardAvoidingView>
  );
}

const styles = StyleSheet.create({
  screen: {
    flex: 1,
    backgroundColor: colors.background,
  },
  content: {
    flexGrow: 1,
    justifyContent: "center",
    paddingHorizontal: 24,
    paddingVertical: 36,
  },
  brandBlock: {
    alignItems: "center",
    marginBottom: 32,
    gap: 8,
  },
  logo: {
    fontSize: 46,
    fontWeight: "900",
    color: colors.text,
    letterSpacing: -1,
  },
  tagline: {
    color: colors.textDim,
    fontSize: 14,
    textAlign: "center",
    maxWidth: 260,
    lineHeight: 18,
  },
  form: {
    gap: 12,
  },
  row: {
    flexDirection: "row",
    gap: 10,
  },
  half: {
    flex: 1,
  },
  input: {
    height: 48,
    borderRadius: 6,
    backgroundColor: colors.cardSubtle,
    borderWidth: 1,
    borderColor: colors.border,
    paddingHorizontal: 14,
    color: colors.text,
    fontSize: 14,
  },
  availabilityRow: {
    marginTop: 4,
    minHeight: 16,
  },
  statusIndicator: {
    flexDirection: "row",
    alignItems: "center",
    gap: 6,
  },
  checkingText: {
    fontSize: 11,
    color: colors.primary,
  },
  statusText: {
    fontSize: 11,
    fontWeight: "600",
  },
  statusAvailable: {
    color: colors.emerald,
  },
  statusUnavailable: {
    color: colors.danger,
  },
  primaryButton: {
    height: 46,
    borderRadius: 8,
    backgroundColor: colors.primary,
    alignItems: "center",
    justifyContent: "center",
    marginTop: 6,
  },
  disabledButton: {
    opacity: 0.45,
  },
  primaryButtonText: {
    color: "#ffffff",
    fontSize: 14.5,
    fontWeight: "700",
  },
  orRow: {
    flexDirection: "row",
    alignItems: "center",
    marginVertical: 24,
    gap: 16,
  },
  orLine: {
    flex: 1,
    height: 0.5,
    backgroundColor: colors.border,
  },
  orText: {
    color: colors.textDim,
    fontSize: 12,
    fontWeight: "700",
  },
  footerWrap: {
    flexDirection: "row",
    justifyContent: "center",
    alignItems: "center",
    paddingVertical: 12,
  },
  footerPrompt: {
    color: colors.textDim,
    fontSize: 13,
  },
  footerAction: {
    color: colors.primary,
    fontSize: 13,
    fontWeight: "700",
  },
  metaBranding: {
    alignItems: "center",
    marginTop: 36,
    gap: 2,
  },
  fromText: {
    color: colors.textDim,
    fontSize: 11,
  },
  brandFooter: {
    color: colors.textMuted,
    fontSize: 12,
    fontWeight: "800",
    letterSpacing: 1.5,
  },
});
