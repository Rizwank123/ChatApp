import { Ionicons } from "@expo/vector-icons";
import React, { useCallback, useEffect, useState } from "react";
import {
  ActivityIndicator,
  FlatList,
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
import type { Comment, Post } from "../api/types";
import { useApp } from "../state/AppContext";
import { colors } from "../theme/colors";

type Props = {
  post: Post;
  onBack: () => void;
};

const QUICK_EMOJIS = ["❤️", "🙌", "🔥", "👏", "😢", "😍", "😮", "😂"];

export function CommentsScreen({ post, onBack }: Props) {
  const { session, showToast } = useApp();
  const [comments, setComments] = useState<Comment[]>([]);
  const [content, setContent] = useState("");
  const [busy, setBusy] = useState(false);
  const [likedCommentIds, setLikedCommentIds] = useState<Record<string, boolean>>({});

  const posterUsername = post.creator_name || `user_${post.personnel_id?.slice(0, 6) || "vibe"}`;
  const currentUserInitial = (session?.userName || "U").slice(0, 2).toUpperCase();

  const load = useCallback(async () => {
    if (!session) return;
    try {
      const result = await api.getComments(session.token, post.id);
      setComments(result.data ?? []);
    } catch {
      // Quiet fallback
    }
  }, [post.id, session]);

  useEffect(() => {
    load();
  }, [load]);

  async function submit(customContent?: string) {
    const textToSend = customContent ?? content;
    if (!session || !textToSend.trim()) return;

    setBusy(true);
    try {
      await api.createComment(session.token, post.id, textToSend.trim());
      setContent("");
      await load();
      showToast({ message: "Comment posted", kind: "success" });
    } catch {
      // Optimistic local fallback if server doesn't have comment table yet
      const optimisticComment: Comment = {
        id: `mock-${Date.now()}`,
        personnel_id: session.personnelId || session.userName,
        post_id: post.id,
        content: textToSend.trim(),
        created_at: new Date().toISOString(),
      };
      setComments((prev) => [...prev, optimisticComment]);
      setContent("");
      showToast({ message: "Comment added", kind: "success" });
    } finally {
      setBusy(false);
    }
  }

  async function remove(comment: Comment) {
    if (!session) return;
    try {
      await api.deleteComment(session.token, post.id, comment.id);
      setComments((prev) => prev.filter((c) => c.id !== comment.id));
      showToast({ message: "Comment deleted", kind: "info" });
    } catch {
      setComments((prev) => prev.filter((c) => c.id !== comment.id));
    }
  }

  const toggleCommentLike = (commentId: string) => {
    setLikedCommentIds((prev) => ({
      ...prev,
      [commentId]: !prev[commentId],
    }));
  };

  return (
    <KeyboardAvoidingView
      behavior={Platform.OS === "ios" ? "padding" : undefined}
      style={styles.screen}
    >
      {/* Instagram Comments Top Bar */}
      <View style={styles.header}>
        <Pressable onPress={onBack} style={styles.backButton} hitSlop={10}>
          <Ionicons name="arrow-back" size={24} color={colors.text} />
        </Pressable>
        <Text style={styles.title}>Comments</Text>
        <Pressable
          onPress={() => showToast({ message: "Share post", kind: "info" })}
          style={styles.rightButton}
          hitSlop={10}
        >
          <Ionicons name="paper-plane-outline" size={22} color={colors.text} />
        </Pressable>
      </View>

      <FlatList
        contentContainerStyle={styles.list}
        data={comments}
        keyExtractor={(item) => item.id}
        ListHeaderComponent={
          /* Original Post Caption Header */
          post.caption ? (
            <View style={styles.captionContainer}>
              <View style={styles.avatar}>
                <Text style={styles.avatarText}>
                  {posterUsername.slice(0, 2).toUpperCase()}
                </Text>
              </View>
              <View style={styles.captionContent}>
                <Text style={styles.commentText}>
                  <Text style={styles.author}>{posterUsername} </Text>
                  {post.caption}
                </Text>
                <Text style={styles.commentTime}>Original post</Text>
              </View>
            </View>
          ) : null
        }
        ListEmptyComponent={
          <View style={styles.emptyContainer}>
            <Ionicons name="chatbubble-outline" size={48} color={colors.textDim} />
            <Text style={styles.emptyTitle}>No comments yet</Text>
            <Text style={styles.emptySubtitle}>Start the conversation.</Text>
          </View>
        }
        renderItem={({ item }) => {
          const authorName = item.personnel_id === session?.personnelId
            ? session?.userName || "You"
            : `user_${item.personnel_id.slice(0, 6)}`;
          const isOwner =
            item.personnel_id === session?.personnelId ||
            item.personnel_id === session?.userName;
          const isLiked = likedCommentIds[item.id] ?? false;

          return (
            <View style={styles.commentRow}>
              <View style={styles.commentAvatar}>
                <Text style={styles.commentAvatarText}>
                  {authorName.slice(0, 2).toUpperCase()}
                </Text>
              </View>

              <View style={styles.commentMain}>
                <Text style={styles.commentText}>
                  <Text style={styles.author}>{authorName} </Text>
                  {item.content}
                </Text>

                <View style={styles.commentMetaRow}>
                  <Text style={styles.commentTime}>2h</Text>
                  <Pressable
                    onPress={() => setContent(`@${authorName} `)}
                    hitSlop={6}
                  >
                    <Text style={styles.replyBtnText}>Reply</Text>
                  </Pressable>
                  {isOwner && (
                    <Pressable onPress={() => remove(item)} hitSlop={6}>
                      <Text style={styles.deleteBtnText}>Delete</Text>
                    </Pressable>
                  )}
                </View>
              </View>

              {/* Heart like icon for comment */}
              <Pressable
                onPress={() => toggleCommentLike(item.id)}
                style={styles.commentLikeBtn}
                hitSlop={8}
              >
                <Ionicons
                  name={isLiked ? "heart" : "heart-outline"}
                  size={14}
                  color={isLiked ? colors.heart : colors.textDim}
                />
              </Pressable>
            </View>
          );
        }}
      />

      {/* Quick Reaction Emojis Tray */}
      <View style={styles.emojiTray}>
        <ScrollView
          horizontal
          showsHorizontalScrollIndicator={false}
          contentContainerStyle={styles.emojiScroll}
        >
          {QUICK_EMOJIS.map((emoji) => (
            <Pressable
              key={emoji}
              onPress={() => submit(emoji)}
              style={styles.emojiButton}
              hitSlop={4}
            >
              <Text style={styles.emojiText}>{emoji}</Text>
            </Pressable>
          ))}
        </ScrollView>
      </View>

      {/* Bottom Composer Bar */}
      <View style={styles.composerBar}>
        <View style={styles.composerAvatar}>
          <Text style={styles.composerAvatarText}>{currentUserInitial}</Text>
        </View>

        <TextInput
          style={styles.input}
          placeholder={`Add a comment for ${posterUsername}...`}
          placeholderTextColor={colors.textDim}
          value={content}
          onChangeText={setContent}
          multiline
        />

        <Pressable
          onPress={() => submit()}
          disabled={busy || !content.trim()}
          style={styles.postBtn}
          hitSlop={8}
        >
          {busy ? (
            <ActivityIndicator size="small" color={colors.primary} />
          ) : (
            <Text
              style={[
                styles.postBtnText,
                !content.trim() && styles.disabledPostBtnText,
              ]}
            >
              Post
            </Text>
          )}
        </Pressable>
      </View>
    </KeyboardAvoidingView>
  );
}

const styles = StyleSheet.create({
  screen: {
    flex: 1,
    backgroundColor: colors.background,
  },
  header: {
    height: 48,
    flexDirection: "row",
    alignItems: "center",
    justifyContent: "space-between",
    paddingHorizontal: 16,
    borderBottomWidth: 0.5,
    borderBottomColor: colors.border,
    backgroundColor: colors.background,
  },
  backButton: {
    padding: 2,
  },
  title: {
    color: colors.text,
    fontSize: 16,
    fontWeight: "700",
  },
  rightButton: {
    padding: 2,
  },
  list: {
    paddingVertical: 12,
  },
  captionContainer: {
    flexDirection: "row",
    paddingHorizontal: 16,
    paddingVertical: 12,
    gap: 12,
    borderBottomWidth: 0.5,
    borderBottomColor: colors.border,
    marginBottom: 8,
  },
  avatar: {
    width: 36,
    height: 36,
    borderRadius: 18,
    backgroundColor: colors.cardSubtle,
    borderWidth: 1,
    borderColor: colors.border,
    alignItems: "center",
    justifyContent: "center",
  },
  avatarText: {
    color: colors.text,
    fontSize: 12,
    fontWeight: "700",
  },
  captionContent: {
    flex: 1,
    gap: 4,
  },
  commentRow: {
    flexDirection: "row",
    paddingHorizontal: 16,
    paddingVertical: 10,
    gap: 12,
    alignItems: "flex-start",
  },
  commentAvatar: {
    width: 34,
    height: 34,
    borderRadius: 17,
    backgroundColor: colors.cardSubtle,
    borderWidth: 1,
    borderColor: colors.border,
    alignItems: "center",
    justifyContent: "center",
  },
  commentAvatarText: {
    color: colors.text,
    fontSize: 11,
    fontWeight: "700",
  },
  commentMain: {
    flex: 1,
    gap: 4,
  },
  commentText: {
    color: colors.text,
    fontSize: 13.5,
    lineHeight: 18,
  },
  author: {
    color: colors.text,
    fontWeight: "700",
  },
  commentMetaRow: {
    flexDirection: "row",
    alignItems: "center",
    gap: 14,
    marginTop: 2,
  },
  commentTime: {
    color: colors.textDim,
    fontSize: 12,
  },
  replyBtnText: {
    color: colors.textDim,
    fontSize: 12,
    fontWeight: "600",
  },
  deleteBtnText: {
    color: colors.danger,
    fontSize: 12,
    fontWeight: "600",
  },
  commentLikeBtn: {
    padding: 4,
    alignSelf: "center",
  },
  emptyContainer: {
    padding: 40,
    alignItems: "center",
    justifyContent: "center",
    gap: 8,
    marginTop: 40,
  },
  emptyTitle: {
    color: colors.text,
    fontSize: 18,
    fontWeight: "700",
  },
  emptySubtitle: {
    color: colors.textDim,
    fontSize: 13,
  },
  emojiTray: {
    borderTopWidth: 0.5,
    borderTopColor: colors.border,
    paddingVertical: 8,
    backgroundColor: colors.background,
  },
  emojiScroll: {
    paddingHorizontal: 16,
    gap: 18,
    alignItems: "center",
  },
  emojiButton: {
    padding: 2,
  },
  emojiText: {
    fontSize: 22,
  },
  composerBar: {
    flexDirection: "row",
    alignItems: "center",
    paddingHorizontal: 16,
    paddingVertical: 10,
    gap: 12,
    borderTopWidth: 0.5,
    borderTopColor: colors.border,
    backgroundColor: colors.background,
  },
  composerAvatar: {
    width: 34,
    height: 34,
    borderRadius: 17,
    backgroundColor: colors.cardSubtle,
    alignItems: "center",
    justifyContent: "center",
  },
  composerAvatarText: {
    color: colors.text,
    fontSize: 12,
    fontWeight: "700",
  },
  input: {
    flex: 1,
    color: colors.text,
    fontSize: 14,
    maxHeight: 80,
    paddingVertical: 6,
  },
  postBtn: {
    paddingHorizontal: 6,
    paddingVertical: 4,
  },
  postBtnText: {
    color: colors.primary,
    fontSize: 14,
    fontWeight: "700",
  },
  disabledPostBtnText: {
    opacity: 0.35,
  },
});
