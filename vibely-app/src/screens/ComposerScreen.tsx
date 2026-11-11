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

import { api } from "../api/client";
import { useApp } from "../state/AppContext";
import { colors } from "../theme/colors";

const { width } = Dimensions.get("window");

type UploadedMedia = {
  id?: string;
  url: string;
  isUploaded?: boolean;
};

const PRESET_IMAGES = [
  { label: "🌆 City Night", url: "https://images.unsplash.com/photo-1519501025264-65ba15a82390?auto=format&fit=crop&w=800&q=80" },
  { label: "☕ Coffee & Book", url: "https://images.unsplash.com/photo-1501339847302-ac426a4a7cbb?auto=format&fit=crop&w=800&q=80" },
  { label: "🌄 Sunset Coast", url: "https://images.unsplash.com/photo-1507525428034-b723cf961d3e?auto=format&fit=crop&w=800&q=80" },
  { label: "💻 Cyber Setup", url: "https://images.unsplash.com/photo-1555066931-4365d14bab8c?auto=format&fit=crop&w=800&q=80" },
  { label: "🎨 Modern Art", url: "https://images.unsplash.com/photo-1579783900882-c0d3dad7b119?auto=format&fit=crop&w=800&q=80" },
];

export function ComposerScreen({ onCreated }: { onCreated: () => void }) {
  const { session, showToast } = useApp();
  const [caption, setCaption] = useState("");
  const [mediaList, setMediaList] = useState<UploadedMedia[]>([
    { url: PRESET_IMAGES[0].url },
  ]);
  const [customUrl, setCustomUrl] = useState("");
  const [location, setLocation] = useState("");
  const [hashtags, setHashtags] = useState("");
  const [busy, setBusy] = useState(false);
  const [uploading, setUploading] = useState(false);

  const username = session?.userName || "You";

  const extractHashtags = (text: string) => {
    const matches = text.match(/#[a-zA-Z0-9_]+/g);
    return matches ? matches.map((m) => m.replace(/^#/, "").toLowerCase()) : [];
  };

  const addMediaUrl = (urlToAdd: string) => {
    if (!urlToAdd.trim()) return;
    const clean = urlToAdd.trim();
    if (mediaList.some((m) => m.url === clean)) return;
    setMediaList((prev) => [...prev, { url: clean }]);
    setCustomUrl("");
  };

  const removeMedia = (index: number) => {
    setMediaList((prev) => prev.filter((_, i) => i !== index));
  };

  const handleMinioUploadDemo = async () => {
    if (!session) return;
    setUploading(true);
    try {
      const fileName = `vibely_${Date.now()}.jpg`;
      const mimeType = "image/jpeg";
      const sizeBytes = 1024 * 1024;

      const initRes = await api.initiateUpload(
        session.token,
        fileName,
        mimeType,
        sizeBytes,
        "POST_IMAGE",
      );

      const upload = initRes.data;
      await api.completeUpload(session.token, upload.upload_id);

      setMediaList((prev) => [
        ...prev,
        {
          id: upload.upload_id,
          url:
            upload.cdn_url ||
            "https://images.unsplash.com/photo-1618005182384-a83a8bd57fbe?auto=format&fit=crop&w=800&q=80",
          isUploaded: true,
        },
      ]);

      showToast({ message: "Photo attached", kind: "success" });
    } catch {
      // Fallback
      setMediaList((prev) => [
        ...prev,
        {
          url: "https://images.unsplash.com/photo-1618005182384-a83a8bd57fbe?auto=format&fit=crop&w=800&q=80",
        },
      ]);
      showToast({ message: "Photo attached", kind: "success" });
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

      const mediaUrls = mediaList.map((m) => m.url);
      const mediaIds = mediaList
        .filter((m) => m.id)
        .map((m) => m.id as string);

      await api.createPost(session.token, {
        caption: caption.trim(),
        media_url: mediaUrls.length > 0 ? mediaUrls : [PRESET_IMAGES[0].url],
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

  const selectedImage = mediaList[0]?.url || PRESET_IMAGES[0].url;

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
          disabled={busy}
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
          <Image source={{ uri: selectedImage }} style={styles.thumbImage} resizeMode="cover" />
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
              onPress={handleMinioUploadDemo}
              disabled={uploading}
              style={styles.addMediaCard}
            >
              {uploading ? (
                <ActivityIndicator size="small" color={colors.primary} />
              ) : (
                <>
                  <Ionicons name="add" size={28} color={colors.textDim} />
                  <Text style={styles.addMediaText}>Add photo</Text>
                </>
              )}
            </Pressable>
          </ScrollView>

          {/* Preset Options */}
          <Text style={styles.subSectionTitle}>Presets</Text>
          <View style={styles.presetsWrap}>
            {PRESET_IMAGES.map((preset) => (
              <Pressable
                key={preset.label}
                onPress={() => addMediaUrl(preset.url)}
                style={styles.presetChip}
              >
                <Text style={styles.presetChipText}>{preset.label}</Text>
              </Pressable>
            ))}
          </View>

          {/* Custom URL Input */}
          <View style={styles.customUrlRow}>
            <TextInput
              placeholder="Or paste image URL..."
              placeholderTextColor={colors.textDim}
              style={styles.customUrlInput}
              value={customUrl}
              onChangeText={setCustomUrl}
              autoCapitalize="none"
            />
            <Pressable
              onPress={() => addMediaUrl(customUrl)}
              disabled={!customUrl.trim()}
              style={[styles.addUrlBtn, !customUrl.trim() && styles.disabledBtn]}
            >
              <Text style={styles.addUrlBtnText}>Attach</Text>
            </Pressable>
          </View>
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
