import { Ionicons } from "@expo/vector-icons";
import React, { useState } from "react";
import {
  ActivityIndicator,
  Dimensions,
  Image,
  KeyboardAvoidingView,
  Platform,
  Pressable,
  ScrollView,
  StyleSheet,
  Text,
  TextInput,
  View,
} from "react-native";
import * as ImagePicker from "expo-image-picker";

import { api } from "../api/client";
import { useApp } from "../state/AppContext";
import { colors } from "../theme/colors";

const { width } = Dimensions.get("window");

type UploadedMedia = {
  id?: string;
  url: string;
  isUploaded?: boolean;
};

export function ComposerScreen({ onCreated }: { onCreated: () => void }) {
  const { session, showToast } = useApp();
  const [caption, setCaption] = useState("");
  const [mediaList, setMediaList] = useState<UploadedMedia[]>([]);
  const [location, setLocation] = useState("");
  const [hashtags, setHashtags] = useState("");
  const [busy, setBusy] = useState(false);
  const [uploading, setUploading] = useState(false);

  const username = session?.userName || "You";

  const extractHashtags = (text: string) => {
    const matches = text.match(/#[a-zA-Z0-9_]+/g);
    return matches ? matches.map((m) => m.replace(/^#/, "").toLowerCase()) : [];
  };

  const removeMedia = (index: number) => {
    setMediaList((prev) => prev.filter((_, i) => i !== index));
  };

  const chooseAndUploadImages = async () => {
    if (!session) return;
    if (mediaList.length >= 10) {
      showToast({ message: "A post can contain up to 10 photos", kind: "info" });
      return;
    }
    setUploading(true);
    try {
      const selection = await ImagePicker.launchImageLibraryAsync({
        mediaTypes: ["images"],
        allowsMultipleSelection: true,
        selectionLimit: 10 - mediaList.length,
        quality: 1,
      });
      if (selection.canceled) return;

      for (const asset of selection.assets) {
        const sourceResponse = await fetch(asset.uri);
        if (!sourceResponse.ok) throw new Error("Could not read the selected photo");
        const blob = await sourceResponse.blob();
        const mimeType = asset.mimeType || blob.type || "image/jpeg";
        const nameExtension = asset.fileName?.includes(".") ? asset.fileName.split(".").pop()?.toLowerCase() : undefined;
        const extension = nameExtension ||
		  ({ "image/jpeg": "jpg", "image/png": "png", "image/webp": "webp", "image/gif": "gif", "image/heic": "heic" } as Record<string, string>)[mimeType];
        if (!extension) throw new Error("This photo format is not supported");
        if (blob.size <= 0 || blob.size > 30 * 1024 * 1024) {
          throw new Error("Each photo must be smaller than 30 MB");
        }

        const fileName = asset.fileName || `vibely_${Date.now()}.${extension}`;
        const initRes = await api.initiateUpload(session.token, fileName, mimeType, blob.size, "POST_IMAGE");
        await api.uploadFileToUrl(initRes.data.upload_url, blob, mimeType);
        await api.completeUpload(session.token, initRes.data.upload_id);
        setMediaList((previous) => [
          ...previous,
          { id: initRes.data.upload_id, url: asset.uri, isUploaded: true },
        ]);
      }

      showToast({ message: "Photos uploaded", kind: "success" });
    } catch (error) {
      showToast({
        message: error instanceof Error ? error.message : "Photo upload failed",
        kind: "error",
      });
    } finally {
      setUploading(false);
    }
  };

  async function create() {
    if (!session) return;
    if (!caption.trim() && mediaList.length === 0) {
      showToast({ message: "Please provide a photo or caption", kind: "error" });
      return;
    }

    setBusy(true);
    try {
      const explicitTags = hashtags
        .split(",")
        .map((t) => t.trim().replace(/^#/, "").toLowerCase())
        .filter(Boolean);
      const inlineTags = extractHashtags(caption);
      const allTags = Array.from(new Set([...explicitTags, ...inlineTags]));

      const mediaIds = mediaList
        .filter((m) => m.id)
        .map((m) => m.id as string);
      const mediaUrls = mediaList.filter((m) => !m.id).map((m) => m.url);

      await api.createPost(session.token, {
        caption: caption.trim(),
        media_url: mediaUrls.length ? mediaUrls : undefined,
        media_ids: mediaIds.length > 0 ? mediaIds : undefined,
        location: location.trim() || undefined,
        visibility: "PUBLIC",
        hashtags: allTags,
      });

      setCaption("");
      setMediaList([]);
      setLocation("");
      setHashtags("");
      showToast({ message: "Post shared to your feed", kind: "success" });
      onCreated();
    } catch (error) {
      showToast({
        message: error instanceof Error ? error.message : "Could not create post",
        kind: "error",
      });
    } finally {
      setBusy(false);
    }
  }

  const selectedImage = mediaList[0]?.url;

  return (
    <KeyboardAvoidingView
      behavior={Platform.OS === "ios" ? "padding" : undefined}
      style={styles.screen}
    >
      {/* Instagram New Post Header */}
      <View style={styles.topHeader}>
        <Text style={styles.headerTitle}>New post</Text>

        <Pressable
          onPress={create}
          disabled={busy || uploading}
          style={styles.shareBtn}
          hitSlop={10}
        >
          {busy ? (
            <ActivityIndicator size="small" color={colors.primary} />
          ) : (
            <Text style={styles.shareBtnText}>Share</Text>
          )}
        </Pressable>
      </View>

      <ScrollView contentContainerStyle={styles.content}>
        {/* Caption & Thumbnail Row (Iconic Instagram style) */}
        <View style={styles.captionRow}>
          {selectedImage ? (
            <Image source={{ uri: selectedImage }} style={styles.thumbImage} resizeMode="cover" />
          ) : (
            <View style={[styles.thumbImage, styles.thumbPlaceholder]}>
              <Ionicons name="image-outline" size={24} color={colors.textDim} />
            </View>
          )}
          <TextInput
            placeholder="Write a caption..."
            placeholderTextColor={colors.textDim}
            multiline
            style={styles.captionInput}
            value={caption}
            onChangeText={setCaption}
          />
        </View>

        <View style={styles.divider} />

        {/* Tag & Settings Rows */}
        <View style={styles.optionsSection}>
          {/* Location row */}
          <View style={styles.optionRow}>
            <Ionicons name="location-outline" size={22} color={colors.text} />
            <TextInput
              placeholder="Add location"
              placeholderTextColor={colors.textDim}
              style={styles.optionInput}
              value={location}
              onChangeText={setLocation}
            />
            {location.length > 0 && (
              <Pressable onPress={() => setLocation("")} hitSlop={6}>
                <Ionicons name="close-circle" size={16} color={colors.textDim} />
              </Pressable>
            )}
          </View>

          {/* Hashtags row */}
          <View style={styles.optionRow}>
            <Ionicons name="pricetag-outline" size={22} color={colors.text} />
            <TextInput
              placeholder="Add hashtags (comma separated: travel, vibely)"
              placeholderTextColor={colors.textDim}
              style={styles.optionInput}
              value={hashtags}
              onChangeText={setHashtags}
              autoCapitalize="none"
            />
          </View>
        </View>

        <View style={styles.divider} />

        {/* Photos & Carousel Selector */}
        <View style={styles.mediaSection}>
          <Text style={styles.sectionTitle}>Select photos</Text>

          {/* Media Carousel Preview */}
          <ScrollView
            horizontal
            showsHorizontalScrollIndicator={false}
            contentContainerStyle={styles.previewScroll}
          >
            {mediaList.map((m, idx) => (
              <View key={idx} style={styles.previewCard}>
                <Image source={{ uri: m.url }} style={styles.previewImage} resizeMode="cover" />
                <Pressable onPress={() => removeMedia(idx)} style={styles.removeMediaBtn} hitSlop={6}>
                  <Ionicons name="close" size={14} color="#ffffff" />
                </Pressable>
                {idx === 0 && (
                  <View style={styles.coverBadge}>
                    <Text style={styles.coverBadgeText}>Cover</Text>
                  </View>
                )}
              </View>
            ))}

            <Pressable
              onPress={chooseAndUploadImages}
              disabled={uploading}
              style={styles.addMediaCard}
            >
              {uploading ? (
                <ActivityIndicator size="small" color={colors.primary} />
              ) : (
                <>
                  <Ionicons name="add" size={28} color={colors.textDim} />
                  <Text style={styles.addMediaText}>Choose photos</Text>
                </>
              )}
            </Pressable>
          </ScrollView>

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
  topHeader: {
    height: 48,
    flexDirection: "row",
    alignItems: "center",
    justifyContent: "space-between",
    paddingHorizontal: 16,
    borderBottomWidth: 0.5,
    borderBottomColor: colors.border,
    backgroundColor: colors.background,
  },
  headerTitle: {
    color: colors.text,
    fontSize: 16,
    fontWeight: "700",
  },
  shareBtn: {
    paddingHorizontal: 4,
    paddingVertical: 4,
  },
  shareBtnText: {
    color: colors.primary,
    fontSize: 16,
    fontWeight: "700",
  },
  content: {
    paddingBottom: 40,
  },
  captionRow: {
    flexDirection: "row",
    padding: 16,
    gap: 12,
    alignItems: "flex-start",
  },
  thumbImage: {
    width: 68,
    height: 68,
    borderRadius: 6,
    backgroundColor: colors.cardSubtle,
  },
  thumbPlaceholder: {
    alignItems: "center",
    justifyContent: "center",
  },
  captionInput: {
    flex: 1,
    color: colors.text,
    fontSize: 14.5,
    lineHeight: 20,
    minHeight: 68,
    textAlignVertical: "top",
  },
  divider: {
    height: 0.5,
    backgroundColor: colors.border,
  },
  optionsSection: {
    paddingVertical: 4,
  },
  optionRow: {
    flexDirection: "row",
    alignItems: "center",
    paddingHorizontal: 16,
    paddingVertical: 12,
    gap: 12,
    borderBottomWidth: 0.5,
    borderBottomColor: colors.border,
  },
  optionInput: {
    flex: 1,
    color: colors.text,
    fontSize: 14.5,
  },
  mediaSection: {
    padding: 16,
  },
  sectionTitle: {
    color: colors.text,
    fontSize: 14,
    fontWeight: "600",
    marginBottom: 12,
  },
  previewScroll: {
    gap: 10,
    paddingBottom: 12,
  },
  previewCard: {
    width: 90,
    height: 90,
    borderRadius: 8,
    overflow: "hidden",
    position: "relative",
    backgroundColor: colors.cardSubtle,
  },
  previewImage: {
    width: "100%",
    height: "100%",
  },
  removeMediaBtn: {
    position: "absolute",
    top: 4,
    right: 4,
    width: 20,
    height: 20,
    borderRadius: 10,
    backgroundColor: "rgba(0,0,0,0.7)",
    alignItems: "center",
    justifyContent: "center",
  },
  coverBadge: {
    position: "absolute",
    bottom: 4,
    left: 4,
    backgroundColor: "rgba(0,0,0,0.7)",
    paddingHorizontal: 5,
    paddingVertical: 2,
    borderRadius: 4,
  },
  coverBadgeText: {
    color: "#ffffff",
    fontSize: 9,
    fontWeight: "700",
  },
  addMediaCard: {
    width: 90,
    height: 90,
    borderRadius: 8,
    borderWidth: 1.5,
    borderStyle: "dashed",
    borderColor: colors.border,
    alignItems: "center",
    justifyContent: "center",
    gap: 4,
    backgroundColor: colors.cardSubtle,
  },
  addMediaText: {
    color: colors.textDim,
    fontSize: 11,
    fontWeight: "600",
  },
  subSectionTitle: {
    color: colors.textDim,
    fontSize: 12,
    fontWeight: "600",
    marginTop: 12,
    marginBottom: 8,
    textTransform: "uppercase",
    letterSpacing: 0.5,
  },
  presetsWrap: {
    flexDirection: "row",
    flexWrap: "wrap",
    gap: 8,
  },
  presetChip: {
    paddingHorizontal: 12,
    paddingVertical: 7,
    borderRadius: 8,
    backgroundColor: colors.cardSubtle,
    borderWidth: 1,
    borderColor: colors.border,
  },
  presetChipText: {
    color: colors.text,
    fontSize: 12.5,
    fontWeight: "500",
  },
  customUrlRow: {
    flexDirection: "row",
    alignItems: "center",
    gap: 10,
    marginTop: 16,
  },
  customUrlInput: {
    flex: 1,
    height: 40,
    borderRadius: 8,
    backgroundColor: colors.cardSubtle,
    borderWidth: 1,
    borderColor: colors.border,
    paddingHorizontal: 12,
    color: colors.text,
    fontSize: 13,
  },
  addUrlBtn: {
    paddingHorizontal: 14,
    height: 40,
    borderRadius: 8,
    backgroundColor: colors.primary,
    alignItems: "center",
    justifyContent: "center",
  },
  disabledBtn: {
    opacity: 0.4,
  },
  addUrlBtnText: {
    color: "#ffffff",
    fontSize: 13,
    fontWeight: "600",
  },
});
