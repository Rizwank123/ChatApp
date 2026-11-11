import { Platform } from "react-native";

type Claims = {
  personnel_id?: string;
  user_id?: string;
  role?: string;
  exp?: number;
  iat?: number;
};

function decodeBase64Url(value: string) {
  const normalized = value.replace(/-/g, "+").replace(/_/g, "/");
  const padded = normalized.padEnd(Math.ceil(normalized.length / 4) * 4, "=");
  if (Platform.OS === "web" && typeof atob === "function") {
    return atob(padded);
  }
  const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/=";
  let output = "";
  let buffer = 0;
  let bits = 0;
  for (const char of padded) {
    if (char === "=") {
      break;
    }
    const valueIndex = chars.indexOf(char);
    if (valueIndex < 0) {
      continue;
    }
    buffer = (buffer << 6) | valueIndex;
    bits += 6;
    if (bits >= 8) {
      bits -= 8;
      output += String.fromCharCode((buffer >> bits) & 0xff);
    }
  }
  return output;
}

export function readJwtClaims(token: string): Claims {
  const payload = token.split(".")[1];
  if (!payload) {
    return {};
  }
  try {
    return JSON.parse(decodeBase64Url(payload)) as Claims;
  } catch {
    return {};
  }
}
