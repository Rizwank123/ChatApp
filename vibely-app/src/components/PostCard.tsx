import { Ionicons } from "@expo/vector-icons";
import React, { useRef, useState } from "react";
import {
  Animated,
  Dimensions,
  Image,
  Modal,
  Pressable,
  StyleSheet,
  Text,
  View,
} from "react-native";

import type { Post } from "../api/types";
import { colors } from "../theme/colors";

const { width } = Dimensions.get("window");

type Props = {
  post: Post;
  onLike: (post: Post) => void;
  onUnlike: (post: Post) => void;
  onComments: (post: Post) => void;
  onSave?: (post: Post) => void;
  onShare?: (post: Post) => void;
  onDelete?: (post: Post) => void;
  onBlockCreator?: (creatorId: string) => void;
  onUserPress?: (username: string, personnelId: string) => void;
};

export function PostCard({
  post,
  onLike,
  onUnlike,
  onComments,
  onSave,
  onShare,
  onDelete,
  onBlockCreator,
  onUserPress,
}: Props) {
  const [isLiked, setIsLiked] = useState(post.is_liked ?? false);
  const [isSaved, setIsSaved] = useState(post.is_saved ?? false);
  const [likeCount, setLikeCount] = useState(post.likes_count ?? post.likes ?? 0);
  const [menuVisible, setMenuVisible] = useState(false);
  const [showHeartAnim, setShowHeartAnim] = useState(false);
  const [activeMediaIndex, setActiveMediaIndex] = useState(0);

  const lastTapRef = useRef<number>(0);
  const heartScale = useRef(new Animated.Value(0)).current;
  const heartOpacity = useRef(new Animated.Value(0)).current;

  const allMedia: string[] =
    post.media && post.media.length > 0
      ? post.media.map((m) => m.url)
      : post.media_url && post.media_url.length > 0
      ? post.media_url
      : [];

  const commentCount = post.comments_count ?? post.comments ?? 0;
  const username = post.creator_name || `user_${post.personnel_id?.slice(0, 6) || "vibe"}`;

  const handleDoubleTap = () => {
    const now = Date.now();
    const DOUBLE_PRESS_DELAY = 300;
    if (now - lastTapRef.current < DOUBLE_PRESS_DELAY) {
      triggerHeartAnimation();
      if (!isLiked) {
        setIsLiked(true);
        setLikeCount((prev) => prev + 1);
        onLike(post);
      }
    }
    lastTapRef.current = now;
  };

  const triggerHeartAnimation = () => {
    setShowHeartAnim(true);
    heartScale.setValue(0.3);
    heartOpacity.setValue(1);

    Animated.parallel([
      Animated.spring(heartScale, {
        toValue: 1.25,
        friction: 3.5,
        useNativeDriver: true,
      }),
      Animated.timing(heartOpacity, {
        toValue: 0,
        duration: 750,
        delay: 200,
        useNativeDriver: true,
      }),
    ]).start(() => {
      setShowHeartAnim(false);
    });
  };

  const toggleLike = () => {
    if (isLiked) {
      setIsLiked(false);
      setLikeCount((prev) => Math.max(0, prev - 1));
      onUnlike(post);
    } else {
      setIsLiked(true);
      setLikeCount((prev) => prev + 1);
      triggerHeartAnimation();
      onLike(post);
    }
  };

  const toggleSave = () => {
    setIsSaved(!isSaved);
    if (onSave) {
      onSave(post);
    }
  };

  const formatTimestamp = (dateStr?: string) => {
    if (!dateStr) return "JUST NOW";
    try {
      const diffMs = Date.now() - new Date(dateStr).getTime();
      const diffMins = Math.floor(diffMs / 60000);
      const diffHours = Math.floor(diffMins / 60);
      const diffDays = Math.floor(diffHours / 24);

      if (diffMins < 1) return "JUST NOW";
      if (diffMins < 60) return `${diffMins} MINUTES AGO`;
      if (diffHours < 24) return `${diffHours} HOURS AGO`;
      if (diffDays < 7) return `${diffDays} DAYS AGO`;

      return new Date(dateStr)
        .toLocaleDateString("en-US", { month: "short", day: "numeric" })
        .toUpperCase();
    } catch {
      return "RECENTLY";
    }
  };

  return (
    <View style={styles.card}>
      {/* Instagram Post Header */}
      <View style={styles.header}>
        <Pressable
          style={styles.headerLeft}
          onPress={() => onUserPress?.(username, post.personnel_id)}
        >
          {/* Story Ring Border around Avatar */}
          <View style={styles.avatarStoryBorder}>
            <View style={styles.avatarInner}>
              <Text style={styles.avatarText}>{username.slice(0, 2).toUpperCase()}</Text>
            </View>
          </View>

          <View style={styles.meta}>
            <Text style={styles.username}>{username}</Text>
            {post.location ? <Text style={styles.location}>{post.location}</Text> : null}
          </View>
        </Pressable>

        <Pressable onPress={() => setMenuVisible(true)} style={styles.moreButton} hitSlop={10}>
          <Ionicons name="ellipsis-horizontal" size={20} color={colors.text} />
        </Pressable>
      </View>

      {/* Media with Double-Tap to Like */}
      <View style={styles.mediaContainer}>
        {allMedia.length > 1 ? (
          <View style={{ width, height: width }}>
            <Animated.ScrollView
              horizontal
              pagingEnabled
              showsHorizontalScrollIndicator={false}
              onMomentumScrollEnd={(e) => {
                const newIndex = Math.round(e.nativeEvent.contentOffset.x / width);
                setActiveMediaIndex(newIndex);
              }}
            >
              {allMedia.map((mUrl, index) => (
                <Pressable
                  key={index}
                  onPress={handleDoubleTap}
                  style={{ width, height: width }}
                >
                  <Image source={{ uri: mUrl }} style={styles.media} resizeMode="cover" />
                </Pressable>
              ))}
            </Animated.ScrollView>

            {/* Top-Right Multi-Image Badge */}
            <View style={styles.badgeIndicator}>
              <Text style={styles.badgeIndicatorText}>
                {activeMediaIndex + 1}/{allMedia.length}
              </Text>
            </View>
          </View>
        ) : (
          <Pressable onPress={handleDoubleTap} style={{ width, height: width }}>
            {allMedia.length === 1 ? (
              <Image source={{ uri: allMedia[0] }} style={styles.media} resizeMode="cover" />
            ) : (
              <View style={styles.placeholderMedia}>
                <Ionicons name="image-outline" size={54} color={colors.textDim} />
                <Text style={styles.placeholderText}>Vibely</Text>
              </View>
            )}
          </Pressable>
        )}

        {/* Big Pop Heart on Double Tap */}
        {showHeartAnim && (
          <Animated.View
            pointerEvents="none"
            style={[
              styles.popHeart,
              {
                opacity: heartOpacity,
                transform: [{ scale: heartScale }],
              },
            ]}
          >
            <Ionicons name="heart" size={100} color={colors.heart} />
          </Animated.View>
        )}
      </View>

      {/* Instagram Action Bar */}
      <View style={styles.actionBar}>
        <View style={styles.leftActions}>
          <Pressable onPress={toggleLike} style={styles.actionIconButton} hitSlop={8}>
            <Ionicons
              name={isLiked ? "heart" : "heart-outline"}
              size={26}
              color={isLiked ? colors.heart : colors.text}
            />
          </Pressable>

          <Pressable onPress={() => onComments(post)} style={styles.actionIconButton} hitSlop={8}>
            <Ionicons name="chatbubble-outline" size={24} color={colors.text} />
          </Pressable>

          <Pressable
            onPress={() => onShare && onShare(post)}
            style={styles.actionIconButton}
            hitSlop={8}
          >
            <Ionicons name="paper-plane-outline" size={24} color={colors.text} />
          </Pressable>
        </View>

        {/* Carousel Pagination Dots */}
        {allMedia.length > 1 && (
          <View style={styles.paginationRow}>
            {allMedia.map((_, i) => (
              <View
                key={i}
                style={[
                  styles.paginationDot,
                  activeMediaIndex === i && styles.paginationDotActive,
                ]}
              />
            ))}
          </View>
        )}

        <Pressable onPress={toggleSave} style={styles.actionIconButton} hitSlop={8}>
          <Ionicons
            name={isSaved ? "bookmark" : "bookmark-outline"}
            size={24}
            color={colors.text}
          />
        </Pressable>
      </View>

      {/* Social Proof & Caption Info */}
      <View style={styles.infoSection}>
        {/* Like Count */}
        <Text style={styles.likeText}>
          {likeCount.toLocaleString()} {likeCount === 1 ? "like" : "likes"}
        </Text>

        {/* Caption */}
        {post.caption ? (
          <View style={styles.captionRow}>
            <Text style={styles.captionText}>
              <Text
                style={styles.captionUser}
                onPress={() => onUserPress?.(username, post.personnel_id)}
              >
                {username}{" "}
              </Text>
              {post.caption}
            </Text>
          </View>
        ) : null}

        {/* Hashtags display */}
        {post.hashtags && post.hashtags.length > 0 && (
          <View style={styles.hashtagsRow}>
            {post.hashtags.map((tag) => (
              <Text key={tag} style={styles.hashtagBadge}>
                #{tag}{" "}
              </Text>
            ))}
          </View>
        )}

        {/* Comments link */}
        {commentCount > 0 ? (
          <Pressable onPress={() => onComments(post)} style={styles.commentsWrap}>
            <Text style={styles.commentsLink}>
              View all {commentCount} {commentCount === 1 ? "comment" : "comments"}
            </Text>
          </Pressable>
        ) : (
          <Pressable onPress={() => onComments(post)} style={styles.commentsWrap}>
            <Text style={styles.commentsLink}>Add a comment...</Text>
          </Pressable>
        )}

        {/* Timestamp */}
        <Text style={styles.timestamp}>{formatTimestamp(post.created_at)}</Text>
      </View>

      {/* Instagram 3-Dots Modal Sheet */}
      <Modal visible={menuVisible} transparent animationType="fade">
        <Pressable style={styles.modalOverlay} onPress={() => setMenuVisible(false)}>
          <View style={styles.modalSheet}>
            <View style={styles.sheetHandle} />

            <Pressable
              style={styles.sheetOption}
              onPress={() => {
                setMenuVisible(false);
                toggleSave();
              }}
            >
              <Ionicons
                name={isSaved ? "bookmark" : "bookmark-outline"}
                size={22}
                color={colors.text}
              />
              <Text style={styles.sheetOptionText}>{isSaved ? "Unsave" : "Save"}</Text>
            </Pressable>

            <Pressable
              style={styles.sheetOption}
              onPress={() => {
                setMenuVisible(false);
                onShare?.(post);
              }}
            >
              <Ionicons name="share-outline" size={22} color={colors.text} />
              <Text style={styles.sheetOptionText}>Share to...</Text>
            </Pressable>

            {onDelete ? (
              <Pressable
                style={styles.sheetOption}
                onPress={() => {
                  setMenuVisible(false);
                  onDelete(post);
                }}
              >
                <Ionicons name="trash-outline" size={22} color={colors.danger} />
                <Text style={[styles.sheetOptionText, { color: colors.danger }]}>
                  Delete Post
                </Text>
              </Pressable>
            ) : null}

            {onBlockCreator ? (
              <Pressable
                style={styles.sheetOption}
                onPress={() => {
                  setMenuVisible(false);
                  onBlockCreator(post.personnel_id);
                }}
              >
                <Ionicons name="hand-left-outline" size={22} color={colors.danger} />
                <Text style={[styles.sheetOptionText, { color: colors.danger }]}>
                  Block User
                </Text>
              </Pressable>
            ) : null}

            <Pressable
              style={[styles.sheetOption, styles.cancelOption]}
              onPress={() => setMenuVisible(false)}
            >
              <Text style={styles.cancelText}>Cancel</Text>
            </Pressable>
          </View>
        </Pressable>
      </Modal>
    </View>
  );
}

const styles = StyleSheet.create({
  card: {
    backgroundColor: colors.card,
    marginBottom: 12,
  },
  header: {
    height: 52,
    flexDirection: "row",
    alignItems: "center",
    justifyContent: "space-between",
    paddingHorizontal: 12,
  },
  headerLeft: {
    flexDirection: "row",
    alignItems: "center",
    gap: 10,
    flex: 1,
  },
  avatarStoryBorder: {
    width: 38,
    height: 38,
    borderRadius: 19,
    padding: 2,
    backgroundColor: colors.storyPink,
    borderWidth: 1,
    borderColor: colors.storyOrange,
    alignItems: "center",
    justifyContent: "center",
  },
  avatarInner: {
    width: 32,
    height: 32,
    borderRadius: 16,
    backgroundColor: colors.cardSubtle,
    alignItems: "center",
    justifyContent: "center",
  },
  avatarText: {
    color: colors.text,
    fontSize: 12,
    fontWeight: "700",
  },
  meta: {
    justifyContent: "center",
  },
  username: {
    color: colors.text,
    fontSize: 13.5,
    fontWeight: "600",
  },
  location: {
    color: colors.textMuted,
    fontSize: 11,
    marginTop: 1,
  },
  moreButton: {
    padding: 6,
  },
  mediaContainer: {
    width: "100%",
    aspectRatio: 1,
    backgroundColor: colors.cardSubtle,
    position: "relative",
    justifyContent: "center",
    alignItems: "center",
  },
  media: {
    width: "100%",
    height: "100%",
  },
  placeholderMedia: {
    flex: 1,
    alignItems: "center",
    justifyContent: "center",
    gap: 8,
    backgroundColor: "#121212",
  },
  placeholderText: {
    color: colors.textDim,
    fontSize: 14,
    fontWeight: "600",
  },
  badgeIndicator: {
    position: "absolute",
    top: 12,
    right: 12,
    backgroundColor: "rgba(0,0,0,0.72)",
    borderRadius: 12,
    paddingHorizontal: 8,
    paddingVertical: 3,
  },
  badgeIndicatorText: {
    color: "#ffffff",
    fontSize: 11,
    fontWeight: "600",
  },
  popHeart: {
    position: "absolute",
    alignSelf: "center",
    zIndex: 10,
    shadowColor: "#000",
    shadowOffset: { width: 0, height: 4 },
    shadowOpacity: 0.3,
    shadowRadius: 8,
  },
  actionBar: {
    flexDirection: "row",
    alignItems: "center",
    justifyContent: "space-between",
    paddingHorizontal: 12,
    paddingTop: 10,
    paddingBottom: 6,
  },
  leftActions: {
    flexDirection: "row",
    alignItems: "center",
    gap: 16,
  },
  actionIconButton: {
    padding: 2,
  },
  paginationRow: {
    flexDirection: "row",
    alignItems: "center",
    gap: 4,
  },
  paginationDot: {
    width: 6,
    height: 6,
    borderRadius: 3,
    backgroundColor: "rgba(255,255,255,0.25)",
  },
  paginationDotActive: {
    backgroundColor: colors.primary,
    width: 7,
    height: 7,
  },
  infoSection: {
    paddingHorizontal: 14,
    paddingTop: 2,
    paddingBottom: 10,
    gap: 4,
  },
  likeText: {
    color: colors.text,
    fontSize: 13.5,
    fontWeight: "700",
  },
  captionRow: {
    marginTop: 2,
  },
  captionText: {
    color: colors.text,
    fontSize: 13.5,
    lineHeight: 18,
  },
  captionUser: {
    fontWeight: "700",
    color: colors.text,
  },
  hashtagsRow: {
    flexDirection: "row",
    flexWrap: "wrap",
    marginTop: 2,
  },
  hashtagBadge: {
    color: colors.primary,
    fontSize: 13,
  },
  commentsWrap: {
    marginTop: 2,
  },
  commentsLink: {
    color: colors.textDim,
    fontSize: 13,
  },
  timestamp: {
    color: colors.textDim,
    fontSize: 10,
    letterSpacing: 0.2,
    marginTop: 2,
  },
  modalOverlay: {
    flex: 1,
    backgroundColor: "rgba(0,0,0,0.65)",
    justifyContent: "flex-end",
  },
  modalSheet: {
    backgroundColor: colors.cardElevated,
    borderTopLeftRadius: 18,
    borderTopRightRadius: 18,
    paddingTop: 12,
    paddingBottom: 28,
    paddingHorizontal: 16,
    gap: 4,
  },
  sheetHandle: {
    width: 40,
    height: 4,
    borderRadius: 2,
    backgroundColor: colors.borderLight,
    alignSelf: "center",
    marginBottom: 12,
  },
  sheetOption: {
    flexDirection: "row",
    alignItems: "center",
    gap: 14,
    paddingVertical: 14,
    paddingHorizontal: 8,
    borderBottomWidth: 0.5,
    borderBottomColor: colors.border,
  },
  sheetOptionText: {
    color: colors.text,
    fontSize: 15,
    fontWeight: "500",
  },
  cancelOption: {
    borderBottomWidth: 0,
    justifyContent: "center",
    paddingTop: 16,
  },
  cancelText: {
    color: colors.textMuted,
    fontSize: 15,
    fontWeight: "600",
    textAlign: "center",
  },
});
