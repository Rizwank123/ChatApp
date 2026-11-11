import React, { useEffect } from "react";
import { StyleSheet, Text } from "react-native";

import { useApp } from "../state/AppContext";
import { colors } from "../theme/colors";

export function Toast() {
  const { toast, clearToast } = useApp();

  useEffect(() => {
    if (!toast) {
      return;
    }
    const timer = setTimeout(clearToast, 3200);
    return () => clearTimeout(timer);
  }, [clearToast, toast]);

  if (!toast) {
    return null;
  }

  return (
    <Text
      style={[
        styles.toast,
        toast.kind === "error" && styles.error,
        toast.kind === "success" && styles.success,
      ]}
    >
      {toast.message}
    </Text>
  );
}

const styles = StyleSheet.create({
  toast: {
    position: "absolute",
    left: 20,
    right: 20,
    bottom: 24,
    zIndex: 999,
    borderRadius: 8,
    paddingVertical: 12,
    paddingHorizontal: 16,
    overflow: "hidden",
    color: "#ffffff",
    backgroundColor: colors.cardElevated,
    borderWidth: 1,
    borderColor: colors.border,
    fontSize: 13.5,
    fontWeight: "600",
    textAlign: "center",
  },
  error: {
    backgroundColor: "rgba(237, 73, 86, 0.95)",
    borderColor: colors.danger,
  },
  success: {
    backgroundColor: "rgba(34, 197, 94, 0.95)",
    borderColor: colors.emerald,
  },
});
