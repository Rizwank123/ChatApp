import { Ionicons } from "@expo/vector-icons";
import React, { useCallback, useEffect, useState } from "react";
import {
  Dimensions,
  FlatList,
  Image,
  Pressable,
  ScrollView,
  StyleSheet,
  Text,
  View,
} from "react-native";

import { api } from "../api/client";
import type { Personnel, Post } from "../api/types";
import { useApp } from "../state/AppContext";
import { colors } from "../theme/colors";

const { width } = Dimensions.get("window");
const GRID_GAP = 1.5;
const GRID_SIZE = (width - GRID_GAP * 2) / 3;

type Tab = "posts" | "reels" | "saved" | "tagged";

type Props = {
  onSelectPost?: (post: Post) => void;
  onLogout?: () => void;
  onOpenFollowList?: (
    initialTab: "followers" | "following",
    followerCount: number,
    followingCount: number,
  ) => void;
  onOpenCreate?: () => void;
};

export function ProfileScreen({
  onSelectPost,
  onLogout,
  onOpenFollowList,
  onOpenCreate,
}: Props) {
  const { session, showToast } = useApp();
  const [profile, setProfile] = useState<Personnel>();
  const [posts, setPosts] = useState<Post[]>([]);
  const [activeTab, setActiveTab] = useState<Tab>("posts");
  const [postCount, setPostCount] = useState(0);
  const [followerCount, setFollowerCount] = useState(0);
  const [followingCount, setFollowingCount] = useState(0);

  const HIGHLIGHTS = [
    { id: "1", title: "Highlights", icon: "sparkles" },
    { id: "2", title: "Travel", icon: "airplane" },
    { id: "3", title: "Art & Code", icon: "code-slash" },
    { id: "4", title: "Memories", icon: "camera" },
  ];

  const loadProfile = useCallback(async () => {
    if (!session?.personnelId) return;
    try {
      const pRes = await api.getPersonnel(session.token, session.personnelId);
      setProfile(pRes.data);

      const postsRes = await api.getUserPosts(session.token, session.personnelId);
      setPosts(postsRes.data ?? []);
      setPostCount(postsRes.total ?? postsRes.data?.length ?? 0);

      const countersRes = await api.getUserCounters(session.token, session.personnelId);
      if (countersRes?.data) {
        setFollowerCount(countersRes.data.follower_count ?? 0);
        setFollowingCount(countersRes.data.following_count ?? 0);
        if (countersRes.data.post_count !== undefined) {
          setPostCount(countersRes.data.post_count);
        }
      }
    } catch {
      // Quiet fallback
    }
  }, [session]);

  useEffect(() => {
    loadProfile();
  }, [loadProfile]);

  const username = session?.userName || "user";
  const displayName =
    profile?.display_name ||
    (profile?.first_name
      ? `${profile.first_name} ${profile.last_name || ""}`.trim()
      : username);

  return (
    <ScrollView style={styles.container} showsVerticalScrollIndicator={false}>
      {/* Instagram Profile Top Bar */}
      <View style={styles.topBar}>
        <View style={styles.topBarLeft}>
          <Text style={styles.topUsername}>{username}</Text>
          <Ionicons name="chevron-down" size={16} color={colors.text} />
        </View>

        <View style={styles.topBarRight}>
          <Pressable style={styles.topIcon} onPress={onOpenCreate} hitSlop={8}>
            <Ionicons name="add-circle-outline" size={26} color={colors.text} />
          </Pressable>
          <Pressable style={styles.topIcon} onPress={onLogout} hitSlop={8}>
            <Ionicons name="log-out-outline" size={24} color={colors.text} />
          </Pressable>
        </View>
      </View>

      {/* Main Profile Header Info */}
      <View style={styles.header}>
        {/* Large 80px Instagram Avatar with Story Ring */}
        <View style={styles.avatarStoryBorder}>
          <View style={styles.avatarInner}>
            <Text style={styles.avatarText}>{username.slice(0, 2).toUpperCase()}</Text>
          </View>
        </View>

        {/* 3 Stats Columns */}
        <View style={styles.statsRow}>
          <View style={styles.statCol}>
            <Text style={styles.statNumber}>{postCount}</Text>
            <Text style={styles.statLabel}>posts</Text>
          </View>
          <Pressable
            style={styles.statCol}
            onPress={() => onOpenFollowList?.("followers", followerCount, followingCount)}
          >
            <Text style={styles.statNumber}>{followerCount}</Text>
            <Text style={styles.statLabel}>followers</Text>
          </Pressable>
          <Pressable
            style={styles.statCol}
            onPress={() => onOpenFollowList?.("following", followerCount, followingCount)}
          >
            <Text style={styles.statNumber}>{followingCount}</Text>
            <Text style={styles.statLabel}>following</Text>
          </Pressable>
        </View>
      </View>

      {/* Bio Section */}
      <View style={styles.bioSection}>
        <Text style={styles.fullName}>{displayName}</Text>
        <Text style={styles.bioText}>
          {profile?.bio || "✨ Living life in vibrant colors. Creating & sharing vibes."}
        </Text>
        <Text style={styles.bioLink}>
          <Ionicons name="link-outline" size={12} color={colors.primary} />{" "}
          vibely.app/@{username}
        </Text>
      </View>

      {/* Profile Action Buttons */}
      <View style={styles.buttonRow}>
        <Pressable
          style={styles.actionBtn}
          onPress={() => showToast({ message: "Profile settings updated", kind: "success" })}
        >
          <Text style={styles.actionBtnText}>Edit profile</Text>
        </Pressable>
        <Pressable
          style={styles.actionBtn}
          onPress={() => showToast({ message: "Profile link copied to clipboard", kind: "success" })}
        >
          <Text style={styles.actionBtnText}>Share profile</Text>
        </Pressable>
        <Pressable
          style={styles.iconActionBtn}
          onPress={() => showToast({ message: "Discover people", kind: "info" })}
        >
          <Ionicons name="person-add-outline" size={18} color={colors.text} />
        </Pressable>
      </View>

      {/* Instagram Story Highlights */}
      <ScrollView
        horizontal
        showsHorizontalScrollIndicator={false}
        contentContainerStyle={styles.highlightsContainer}
      >
        {HIGHLIGHTS.map((h) => (
          <View key={h.id} style={styles.highlightCol}>
            <View style={styles.highlightRing}>
              <View style={styles.highlightInner}>
                <Ionicons name={h.icon as any} size={22} color={colors.text} />
              </View>
            </View>
            <Text style={styles.highlightText} numberOfLines={1}>
              {h.title}
            </Text>
          </View>
        ))}

        {/* Add Highlight Button */}
        <Pressable
          style={styles.highlightCol}
          onPress={() => showToast({ message: "Create new story highlight", kind: "info" })}
        >
          <View style={styles.newHighlightRing}>
            <Ionicons name="add" size={26} color={colors.text} />
          </View>
          <Text style={styles.highlightText}>New</Text>
        </Pressable>
      </ScrollView>

      {/* Instagram Tab Navigation */}
      <View style={styles.tabBar}>
        <Pressable
          style={[styles.tabButton, activeTab === "posts" && styles.activeTab]}
          onPress={() => setActiveTab("posts")}
        >
          <Ionicons
            name="grid-outline"
            size={22}
            color={activeTab === "posts" ? colors.text : colors.textDim}
          />
        </Pressable>
        <Pressable
          style={[styles.tabButton, activeTab === "reels" && styles.activeTab]}
          onPress={() => setActiveTab("reels")}
        >
          <Ionicons
            name="play-outline"
            size={22}
            color={activeTab === "reels" ? colors.text : colors.textDim}
          />
        </Pressable>
        <Pressable
          style={[styles.tabButton, activeTab === "saved" && styles.activeTab]}
          onPress={() => setActiveTab("saved")}
        >
          <Ionicons
            name="bookmark-outline"
            size={22}
            color={activeTab === "saved" ? colors.text : colors.textDim}
          />
        </Pressable>
        <Pressable
          style={[styles.tabButton, activeTab === "tagged" && styles.activeTab]}
          onPress={() => setActiveTab("tagged")}
        >
          <Ionicons
            name="person-outline"
            size={22}
            color={activeTab === "tagged" ? colors.text : colors.textDim}
          />
        </Pressable>
      </View>

      {/* 3-Column Posts Grid */}
      <View style={styles.gridContainer}>
        {posts.map((post) => {
          const media = post.media_url?.[0] || post.media?.[0]?.url;
          return (
            <Pressable
              key={post.id}
              onPress={() => onSelectPost?.(post)}
              style={styles.gridItem}
            >
              {media ? (
                <Image source={{ uri: media }} style={styles.gridMedia} resizeMode="cover" />
              ) : (
                <View style={styles.gridPlaceholder}>
                  <Ionicons name="image-outline" size={24} color={colors.textDim} />
                  <Text style={styles.gridCaption} numberOfLines={2}>
                    {post.caption || "Post"}
                  </Text>
                </View>
              )}
            </Pressable>
          );
        })}
        {posts.length === 0 && (
          <View style={styles.emptyGrid}>
            <View style={styles.emptyIconCircle}>
              <Ionicons name="camera-outline" size={48} color={colors.text} />
            </View>
            <Text style={styles.emptyTitle}>Share Photos</Text>
            <Text style={styles.emptySubtitle}>
              When you share photos, they will appear on your profile.
            </Text>
            <Pressable onPress={onOpenCreate} style={styles.shareFirstBtn}>
              <Text style={styles.shareFirstBtnText}>Share your first photo</Text>
            </Pressable>
          </View>
        )}
      </View>
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: colors.background,
  },
  topBar: {
    height: 48,
    flexDirection: "row",
    alignItems: "center",
    justifyContent: "space-between",
    paddingHorizontal: 16,
    backgroundColor: colors.background,
  },
  topBarLeft: {
    flexDirection: "row",
    alignItems: "center",
    gap: 6,
  },
  topUsername: {
    color: colors.text,
    fontSize: 20,
    fontWeight: "700",
  },
  topBarRight: {
    flexDirection: "row",
    alignItems: "center",
    gap: 16,
  },
  topIcon: {
    padding: 2,
  },
  header: {
    flexDirection: "row",
    alignItems: "center",
    paddingHorizontal: 16,
    paddingTop: 12,
    paddingBottom: 6,
    gap: 24,
  },
  avatarStoryBorder: {
    width: 84,
    height: 84,
    borderRadius: 42,
    borderWidth: 2,
    borderColor: colors.storyOrange,
    padding: 2.5,
    alignItems: "center",
    justifyContent: "center",
  },
  avatarInner: {
    width: 74,
    height: 74,
    borderRadius: 37,
    backgroundColor: colors.cardSubtle,
    alignItems: "center",
    justifyContent: "center",
  },
  avatarText: {
    color: colors.text,
    fontSize: 24,
    fontWeight: "700",
  },
  statsRow: {
    flex: 1,
    flexDirection: "row",
    justifyContent: "space-around",
    alignItems: "center",
  },
  statCol: {
    alignItems: "center",
    gap: 2,
  },
  statNumber: {
    color: colors.text,
    fontSize: 18,
    fontWeight: "700",
  },
  statLabel: {
    color: colors.text,
    fontSize: 13,
  },
  bioSection: {
    paddingHorizontal: 16,
    paddingTop: 8,
    paddingBottom: 14,
    gap: 3,
  },
  fullName: {
    color: colors.text,
    fontSize: 14,
    fontWeight: "700",
  },
  bioText: {
    color: colors.text,
    fontSize: 13.5,
    lineHeight: 18,
  },
  bioLink: {
    color: colors.primary,
    fontSize: 13.5,
    fontWeight: "500",
    marginTop: 2,
  },
  buttonRow: {
    flexDirection: "row",
    paddingHorizontal: 16,
    gap: 8,
    marginBottom: 16,
  },
  actionBtn: {
    flex: 1,
    height: 34,
    borderRadius: 8,
    backgroundColor: colors.cardSubtle,
    borderWidth: 1,
    borderColor: colors.border,
    alignItems: "center",
    justifyContent: "center",
  },
  actionBtnText: {
    color: colors.text,
    fontSize: 13.5,
    fontWeight: "600",
  },
  iconActionBtn: {
    width: 34,
    height: 34,
    borderRadius: 8,
    backgroundColor: colors.cardSubtle,
    borderWidth: 1,
    borderColor: colors.border,
    alignItems: "center",
    justifyContent: "center",
  },
  highlightsContainer: {
    paddingHorizontal: 16,
    gap: 16,
    paddingBottom: 16,
  },
  highlightCol: {
    alignItems: "center",
    width: 64,
    gap: 5,
  },
  highlightRing: {
    width: 64,
    height: 64,
    borderRadius: 32,
    borderWidth: 1,
    borderColor: colors.border,
    alignItems: "center",
    justifyContent: "center",
    padding: 2,
  },
  highlightInner: {
    width: 56,
    height: 56,
    borderRadius: 28,
    backgroundColor: colors.cardSubtle,
    alignItems: "center",
    justifyContent: "center",
  },
  newHighlightRing: {
    width: 64,
    height: 64,
    borderRadius: 32,
    borderWidth: 1,
    borderColor: colors.border,
    alignItems: "center",
    justifyContent: "center",
    backgroundColor: colors.cardSubtle,
  },
  highlightText: {
    color: colors.text,
    fontSize: 11.5,
    textAlign: "center",
  },
  tabBar: {
    flexDirection: "row",
    borderTopWidth: 0.5,
    borderTopColor: colors.border,
    backgroundColor: colors.background,
  },
  tabButton: {
    flex: 1,
    height: 44,
    alignItems: "center",
    justifyContent: "center",
  },
  activeTab: {
    borderBottomWidth: 1.5,
    borderBottomColor: colors.text,
  },
  gridContainer: {
    flexDirection: "row",
    flexWrap: "wrap",
    paddingBottom: 24,
  },
  gridItem: {
    width: GRID_SIZE,
    height: GRID_SIZE,
    marginRight: GRID_GAP,
    marginBottom: GRID_GAP,
    backgroundColor: colors.cardSubtle,
  },
  gridMedia: {
    width: "100%",
    height: "100%",
  },
  gridPlaceholder: {
    flex: 1,
    alignItems: "center",
    justifyContent: "center",
    padding: 6,
    gap: 4,
  },
  gridCaption: {
    color: colors.textDim,
    fontSize: 10,
    textAlign: "center",
  },
  emptyGrid: {
    width: "100%",
    padding: 40,
    alignItems: "center",
    justifyContent: "center",
    gap: 10,
    marginTop: 20,
  },
  emptyIconCircle: {
    width: 80,
    height: 80,
    borderRadius: 40,
    borderWidth: 2,
    borderColor: colors.text,
    alignItems: "center",
    justifyContent: "center",
    marginBottom: 8,
  },
  emptyTitle: {
    color: colors.text,
    fontSize: 22,
    fontWeight: "700",
  },
  emptySubtitle: {
    color: colors.textMuted,
    fontSize: 14,
    textAlign: "center",
    maxWidth: 260,
    lineHeight: 18,
  },
  shareFirstBtn: {
    marginTop: 8,
  },
  shareFirstBtnText: {
    color: colors.primary,
    fontSize: 14,
    fontWeight: "600",
  },
});
