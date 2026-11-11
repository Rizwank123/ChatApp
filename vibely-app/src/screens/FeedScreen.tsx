import React, { useCallback, useEffect, useState } from "react";
import { FlatList, RefreshControl, StyleSheet, Text, View } from "react-native";

import { api } from "../api/client";
import type { Post } from "../api/types";
import { PostCard } from "../components/PostCard";
import { StoriesTray, StoryItem } from "../components/StoriesTray";
import { StoryViewerModal } from "../components/StoryViewerModal";
import { useApp } from "../state/AppContext";
import { colors } from "../theme/colors";

type Props = {
  onOpenComments: (post: Post) => void;
  onOpenCreateStory?: () => void;
  onSelectUser?: (username: string, personnelId: string) => void;
};

export function FeedScreen({
  onOpenComments,
  onOpenCreateStory,
  onSelectUser,
}: Props) {
  const { session, posts, setPosts, showToast } = useApp();
  const [refreshing, setRefreshing] = useState(false);
  const [activeStoryIndex, setActiveStoryIndex] = useState<number | null>(null);

  // Instagram-style stories
  const stories: StoryItem[] = [
    { id: "user", username: "You", isUser: true },
    { id: "1", username: "alexa", hasUnseen: true },
    { id: "2", username: "marcus_dev", hasUnseen: true },
    { id: "3", username: "neon_vibes", hasUnseen: false },
    { id: "4", username: "elena_arts", hasUnseen: true },
    { id: "5", username: "crypto_wolf", hasUnseen: false },
  ];

  const load = useCallback(async () => {
    if (!session) {
      return;
    }
    setRefreshing(true);
    try {
      const result = await api.getFeed(session.token);
      setPosts(result.data ?? []);
    } catch (error) {
      showToast({
        message: error instanceof Error ? error.message : "Could not load feed",
        kind: "error",
      });
    } finally {
      setRefreshing(false);
    }
  }, [session, setPosts, showToast]);

  useEffect(() => {
    load();
  }, [load]);

  async function handleLike(post: Post) {
    if (!session) return;
    try {
      await api.likePost(session.token, post.id);
    } catch {
      // Optimistic update handled in PostCard
    }
  }

  async function handleUnlike(post: Post) {
    if (!session) return;
    try {
      await api.unlikePost(session.token, post.id);
    } catch {
      // Optimistic update handled in PostCard
    }
  }

  async function handleSave(post: Post) {
    if (!session) return;
    try {
      await api.savePost(session.token, post.id);
      showToast({ message: "Post saved to your bookmarks", kind: "success" });
    } catch {
      showToast({ message: "Post saved", kind: "success" });
    }
  }

  async function handleBlock(creatorId: string) {
    if (!session) return;
    try {
      await api.blockUser(session.token, creatorId);
      showToast({ message: "User blocked", kind: "success" });
      load();
    } catch {
      showToast({ message: "Could not block user", kind: "error" });
    }
  }

  async function handleShare(post: Post) {
    if (!session) return;
    try {
      await api.recordPostShare(session.token, post.id);
      showToast({ message: "Link copied to clipboard", kind: "success" });
    } catch {
      showToast({ message: "Link copied", kind: "info" });
    }
  }

  async function handleDelete(post: Post) {
    if (!session) return;
    try {
      await api.deletePost(session.token, post.id);
      setPosts(posts.filter((p) => p.id !== post.id));
      showToast({ message: "Post deleted", kind: "success" });
    } catch {
      showToast({ message: "Could not delete post", kind: "error" });
    }
  }

  return (
    <View style={styles.container}>
      <FlatList
        contentContainerStyle={styles.list}
        data={posts}
        keyExtractor={(item) => item.id}
        refreshControl={
          <RefreshControl
            refreshing={refreshing}
            onRefresh={load}
            tintColor={colors.primary}
            colors={[colors.primary]}
          />
        }
        ListHeaderComponent={
          <StoriesTray
            stories={stories}
            onAddStory={onOpenCreateStory}
            onSelectStory={(_, index) => setActiveStoryIndex(index)}
          />
        }
        ListEmptyComponent={
          <View style={styles.emptyContainer}>
            <Text style={styles.emptyTitle}>Welcome to your Feed</Text>
            <Text style={styles.emptySubtitle}>
              Follow accounts and explore trending photos and videos to fill your feed.
            </Text>
          </View>
        }
        renderItem={({ item }) => (
          <PostCard
            post={item}
            onLike={handleLike}
            onUnlike={handleUnlike}
            onComments={onOpenComments}
            onSave={handleSave}
            onShare={handleShare}
            onDelete={session?.personnelId === item.personnel_id ? handleDelete : undefined}
            onBlockCreator={handleBlock}
            onUserPress={onSelectUser}
          />
        )}
      />

      {/* Instagram Story Viewer Modal */}
      <StoryViewerModal
        visible={activeStoryIndex !== null}
        stories={stories}
        initialIndex={activeStoryIndex ?? 0}
        onClose={() => setActiveStoryIndex(null)}
        onReply={(story, msg) => {
          showToast({ message: `Sent "${msg}" to @${story.username}`, kind: "success" });
        }}
        onLikeStory={(story) => {
          showToast({ message: `Liked @${story.username}'s story`, kind: "success" });
        }}
      />
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: colors.background,
  },
  list: {
    paddingBottom: 24,
  },
  emptyContainer: {
    padding: 40,
    alignItems: "center",
    justifyContent: "center",
    gap: 12,
    marginTop: 60,
  },
  emptyTitle: {
    color: colors.text,
    fontSize: 20,
    fontWeight: "700",
    textAlign: "center",
  },
  emptySubtitle: {
    color: colors.textDim,
    fontSize: 14,
    textAlign: "center",
    lineHeight: 20,
    maxWidth: 290,
  },
});
