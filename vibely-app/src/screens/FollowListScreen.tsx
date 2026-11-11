import { Ionicons } from "@expo/vector-icons";
import React, { useCallback, useEffect, useRef, useState } from "react";
import {
  ActivityIndicator,
  FlatList,
  Image,
  Pressable,
  StyleSheet,
  Text,
  TextInput,
  View,
} from "react-native";

import { api } from "../api/client";
import type { FollowUserItem, UUID } from "../api/types";
import { useApp } from "../state/AppContext";
import { colors } from "../theme/colors";

type TabType = "followers" | "following";

type Props = {
  targetPersonnelId: UUID;
  targetUsername?: string;
  initialTab?: TabType;
  followerCount?: number;
  followingCount?: number;
  onBack: () => void;
  onOpenChat: (participantId: UUID, participantName: string) => void;
};

const PAGE_SIZE = 20;

export function FollowListScreen({
  targetPersonnelId,
  targetUsername = "User",
  initialTab = "followers",
  followerCount,
  followingCount,
  onBack,
  onOpenChat,
}: Props) {
  const { session, showToast } = useApp();
  const [activeTab, setActiveTab] = useState<TabType>(initialTab);
  const [searchQuery, setSearchQuery] = useState("");
  const [debouncedQuery, setDebouncedQuery] = useState("");

  // Followers state
  const [followers, setFollowers] = useState<FollowUserItem[]>([]);
  const [followersPage, setFollowersPage] = useState(0);
  const [followersTotal, setFollowersTotal] = useState<number | null>(followerCount ?? null);
  const [hasMoreFollowers, setHasMoreFollowers] = useState(true);
  const [loadingFollowers, setLoadingFollowers] = useState(false);
  const [refreshingFollowers, setRefreshingFollowers] = useState(false);

  // Following state
  const [following, setFollowing] = useState<FollowUserItem[]>([]);
  const [followingPage, setFollowingPage] = useState(0);
  const [followingTotal, setFollowingTotal] = useState<number | null>(followingCount ?? null);
  const [hasMoreFollowing, setHasMoreFollowing] = useState(true);
  const [loadingFollowing, setLoadingFollowing] = useState(false);
  const [refreshingFollowing, setRefreshingFollowing] = useState(false);

  // Action states
  const [togglingFollow, setTogglingFollow] = useState<Record<string, boolean>>({});
  const [checkingChat, setCheckingChat] = useState<Record<string, boolean>>({});

  // Debounce search query
  useEffect(() => {
    const handler = setTimeout(() => {
      setDebouncedQuery(searchQuery.trim());
    }, 280);
    return () => clearTimeout(handler);
  }, [searchQuery]);

  // Load followers
  const loadFollowers = useCallback(
    async (pageToLoad: number, queryToUse: string, isRefresh = false) => {
      if (!session) return;
      if (isRefresh) {
        setRefreshingFollowers(true);
      } else {
        setLoadingFollowers(true);
      }

      try {
        const res = await api.getFollowers(
          session.token,
          targetPersonnelId,
          pageToLoad,
          PAGE_SIZE,
          queryToUse,
        );
        const list = res.data ?? [];
        const total = res.total ?? 0;

        setFollowers((prev) => (pageToLoad === 0 ? list : [...prev, ...list]));
        setFollowersPage(pageToLoad);
        setFollowersTotal(total);
        setHasMoreFollowers(list.length === PAGE_SIZE && (pageToLoad + 1) * PAGE_SIZE < total);
      } catch (err) {
        showToast({
          message: err instanceof Error ? err.message : "Failed to load followers",
          kind: "error",
        });
      } finally {
        setLoadingFollowers(false);
        setRefreshingFollowers(false);
      }
    },
    [session, showToast, targetPersonnelId],
  );

  // Load following
  const loadFollowing = useCallback(
    async (pageToLoad: number, queryToUse: string, isRefresh = false) => {
      if (!session) return;
      if (isRefresh) {
        setRefreshingFollowing(true);
      } else {
        setLoadingFollowing(true);
      }

      try {
        const res = await api.getFollowing(
          session.token,
          targetPersonnelId,
          pageToLoad,
          PAGE_SIZE,
          queryToUse,
        );
        const list = res.data ?? [];
        const total = res.total ?? 0;

        setFollowing((prev) => (pageToLoad === 0 ? list : [...prev, ...list]));
        setFollowingPage(pageToLoad);
        setFollowingTotal(total);
        setHasMoreFollowing(list.length === PAGE_SIZE && (pageToLoad + 1) * PAGE_SIZE < total);
      } catch (err) {
        showToast({
          message: err instanceof Error ? err.message : "Failed to load following",
          kind: "error",
        });
      } finally {
        setLoadingFollowing(false);
        setRefreshingFollowing(false);
      }
    },
    [session, showToast, targetPersonnelId],
  );

  // Trigger loads when tab or query changes
  useEffect(() => {
    if (activeTab === "followers") {
      loadFollowers(0, debouncedQuery);
    } else {
      loadFollowing(0, debouncedQuery);
    }
  }, [activeTab, debouncedQuery, loadFollowers, loadFollowing]);

  const handleEndReached = () => {
    if (activeTab === "followers") {
      if (!loadingFollowers && hasMoreFollowers) {
        loadFollowers(followersPage + 1, debouncedQuery);
      }
    } else {
      if (!loadingFollowing && hasMoreFollowing) {
        loadFollowing(followingPage + 1, debouncedQuery);
      }
    }
  };

  const handleRefresh = () => {
    if (activeTab === "followers") {
      loadFollowers(0, debouncedQuery, true);
    } else {
      loadFollowing(0, debouncedQuery, true);
    }
  };

  // Follow / unfollow handler
  const handleToggleFollow = async (item: FollowUserItem) => {
    if (!session || togglingFollow[item.personnel_id]) return;

    setTogglingFollow((prev) => ({ ...prev, [item.personnel_id]: true }));
    const wasFollowing = item.is_following;
    const newIsFollowing = !wasFollowing;

    // Optimistically update both lists
    const updater = (list: FollowUserItem[]) =>
      list.map((u) =>
        u.personnel_id === item.personnel_id
          ? {
              ...u,
              is_following: newIsFollowing,
              is_mutual: newIsFollowing && u.follows_you,
            }
          : u,
      );

    setFollowers(updater);
    setFollowing(updater);

    try {
      if (wasFollowing) {
        await api.unfollowSocial(session.token, item.personnel_id);
        showToast({ message: `Unfollowed @${item.user_name}`, kind: "info" });
      } else {
        await api.followSocial(session.token, item.personnel_id);
        showToast({ message: `Following @${item.user_name}`, kind: "success" });
      }
    } catch (err) {
      // Revert on failure
      const reverter = (list: FollowUserItem[]) =>
        list.map((u) =>
          u.personnel_id === item.personnel_id
            ? {
                ...u,
                is_following: wasFollowing,
                is_mutual: wasFollowing && u.follows_you,
              }
            : u,
        );
      setFollowers(reverter);
      setFollowing(reverter);

      showToast({
        message: err instanceof Error ? err.message : "Follow action failed",
        kind: "error",
      });
    } finally {
      setTogglingFollow((prev) => ({ ...prev, [item.personnel_id]: false }));
    }
  };

  // Direct message authorization check and navigation
  const handleUserClick = async (item: FollowUserItem) => {
    if (!session || checkingChat[item.personnel_id]) return;

    // Don't chat with self
    if (item.personnel_id === session.personnelId) {
      showToast({ message: "You cannot message yourself.", kind: "info" });
      return;
    }

    setCheckingChat((prev) => ({ ...prev, [item.personnel_id]: true }));

    try {
      const res = await api.canChat(session.token, item.personnel_id);
      const data = res.data;

      if (data?.can_chat) {
        onOpenChat(item.personnel_id, item.display_name || item.user_name);
      } else {
        const reason =
          data?.reason ||
          "Messaging is restricted: either you must follow this user or they must follow you.";
        showToast({
          message: `⛔ ${reason}`,
          kind: "error",
        });
      }
    } catch (err) {
      const msg = err instanceof Error ? err.message : "Failed to verify chat permission";
      showToast({ message: msg, kind: "error" });
    } finally {
      setCheckingChat((prev) => ({ ...prev, [item.personnel_id]: false }));
    }
  };

  const currentList = activeTab === "followers" ? followers : following;
  const isLoading = activeTab === "followers" ? loadingFollowers : loadingFollowing;
  const isRefreshing = activeTab === "followers" ? refreshingFollowers : refreshingFollowing;

  const renderItem = ({ item }: { item: FollowUserItem }) => {
    const isSelf = item.personnel_id === session?.personnelId;
    const isFollowPending = togglingFollow[item.personnel_id];
    const isChatPending = checkingChat[item.personnel_id];
    const initials = (item.user_name || "U").slice(0, 2).toUpperCase();

    return (
      <Pressable
        style={styles.userRow}
        onPress={() => handleUserClick(item)}
        android_ripple={{ color: colors.border }}
      >
        {/* Avatar */}
        <View style={styles.avatarWrap}>
          {item.avatar ? (
            <Image source={{ uri: item.avatar }} style={styles.avatarImage} />
          ) : (
            <View style={styles.avatarPlaceholder}>
              <Text style={styles.avatarInitials}>{initials}</Text>
            </View>
          )}
        </View>

        {/* User Info */}
        <View style={styles.userInfo}>
          <View style={styles.nameRow}>
            <Text style={styles.displayName} numberOfLines={1}>
              {item.display_name || item.user_name}
            </Text>
            {item.is_mutual && (
              <View style={styles.mutualBadge}>
                <Text style={styles.mutualBadgeText}>Mutual</Text>
              </View>
            )}
          </View>
          <Text style={styles.username} numberOfLines={1}>
            @{item.user_name}
          </Text>
          {item.bio ? (
            <Text style={styles.bioText} numberOfLines={1}>
              {item.bio}
            </Text>
          ) : null}
        </View>

        {/* Action Buttons */}
        {!isSelf && (
          <View style={styles.actions}>
            {/* Direct Message button */}
            <Pressable
              style={styles.messageBtn}
              onPress={() => handleUserClick(item)}
              disabled={isChatPending}
            >
              {isChatPending ? (
                <ActivityIndicator size="small" color={colors.text} />
              ) : (
                <Ionicons name="chatbubble-ellipses-outline" size={17} color={colors.text} />
              )}
            </Pressable>

            {/* Follow / Unfollow button */}
            <Pressable
              style={[
                styles.followBtn,
                item.is_following ? styles.followingBtn : styles.notFollowingBtn,
              ]}
              onPress={() => handleToggleFollow(item)}
              disabled={isFollowPending}
            >
              {isFollowPending ? (
                <ActivityIndicator
                  size="small"
                  color={item.is_following ? colors.text : "#000000"}
                />
              ) : (
                <Text
                  style={[
                    styles.followBtnText,
                    item.is_following ? styles.followingBtnText : styles.notFollowingBtnText,
                  ]}
                >
                  {item.is_following ? "Following" : "Follow"}
                </Text>
              )}
            </Pressable>
          </View>
        )}
      </Pressable>
    );
  };

  return (
    <View style={styles.container}>
      {/* Top Bar Header */}
      <View style={styles.header}>
        <Pressable onPress={onBack} style={styles.backButton}>
          <Ionicons name="arrow-back" size={24} color={colors.text} />
        </Pressable>
        <Text style={styles.headerTitle} numberOfLines={1}>
          {targetUsername}
        </Text>
        <View style={{ width: 28 }} />
      </View>

      {/* Tabs Row */}
      <View style={styles.tabRow}>
        <Pressable
          style={[styles.tab, activeTab === "followers" && styles.activeTab]}
          onPress={() => setActiveTab("followers")}
        >
          <Text style={[styles.tabText, activeTab === "followers" && styles.activeTabText]}>
            Followers {followersTotal !== null ? `(${followersTotal})` : ""}
          </Text>
        </Pressable>
        <Pressable
          style={[styles.tab, activeTab === "following" && styles.activeTab]}
          onPress={() => setActiveTab("following")}
        >
          <Text style={[styles.tabText, activeTab === "following" && styles.activeTabText]}>
            Following {followingTotal !== null ? `(${followingTotal})` : ""}
          </Text>
        </Pressable>
      </View>

      {/* Real-time Search Filter Bar */}
      <View style={styles.searchContainer}>
        <Ionicons name="search" size={17} color={colors.textDim} style={styles.searchIcon} />
        <TextInput
          style={styles.searchInput}
          placeholder={`Search ${activeTab}...`}
          placeholderTextColor={colors.textDim}
          value={searchQuery}
          onChangeText={setSearchQuery}
          autoCapitalize="none"
          autoCorrect={false}
          clearButtonMode="while-editing"
        />
        {searchQuery.length > 0 && (
          <Pressable onPress={() => setSearchQuery("")} style={styles.clearBtn}>
            <Ionicons name="close-circle" size={17} color={colors.textDim} />
          </Pressable>
        )}
      </View>

      {/* Paginated User List */}
      <FlatList
        data={currentList}
        keyExtractor={(item) => item.personnel_id}
        renderItem={renderItem}
        contentContainerStyle={styles.listContent}
        onEndReached={handleEndReached}
        onEndReachedThreshold={0.3}
        refreshing={isRefreshing}
        onRefresh={handleRefresh}
        ListFooterComponent={
          isLoading && currentList.length > 0 ? (
            <View style={styles.loadingFooter}>
              <ActivityIndicator size="small" color={colors.cyan} />
            </View>
          ) : null
        }
        ListEmptyComponent={
          !isLoading ? (
            <View style={styles.emptyContainer}>
              <Ionicons
                name={activeTab === "followers" ? "people-outline" : "person-add-outline"}
                size={48}
                color={colors.textDim}
              />
              <Text style={styles.emptyTitle}>
                {searchQuery
                  ? `No ${activeTab} match "${searchQuery}"`
                  : activeTab === "followers"
                  ? "No followers yet"
                  : "Not following anyone yet"}
              </Text>
              <Text style={styles.emptySubtitle}>
                {activeTab === "followers"
                  ? "When users follow this account, they will appear here."
                  : "Users this account follows will appear here."}
              </Text>
            </View>
          ) : (
            <View style={styles.loadingContainer}>
              <ActivityIndicator size="large" color={colors.cyan} />
            </View>
          )
        }
      />
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: colors.background,
  },
  header: {
    height: 52,
    flexDirection: "row",
    alignItems: "center",
    justifyContent: "space-between",
    paddingHorizontal: 16,
    backgroundColor: colors.canvas,
    borderBottomWidth: 1,
    borderBottomColor: colors.border,
  },
  backButton: {
    padding: 4,
  },
  headerTitle: {
    fontSize: 17,
    fontWeight: "700",
    color: colors.text,
  },
  tabRow: {
    flexDirection: "row",
    backgroundColor: colors.canvas,
    borderBottomWidth: 1,
    borderBottomColor: colors.border,
  },
  tab: {
    flex: 1,
    paddingVertical: 13,
    alignItems: "center",
    justifyContent: "center",
    borderBottomWidth: 2,
    borderBottomColor: "transparent",
  },
  activeTab: {
    borderBottomColor: colors.text,
  },
  tabText: {
    fontSize: 14,
    fontWeight: "600",
    color: colors.textDim,
  },
  activeTabText: {
    color: colors.text,
    fontWeight: "700",
  },
  searchContainer: {
    flexDirection: "row",
    alignItems: "center",
    marginHorizontal: 16,
    marginVertical: 10,
    paddingHorizontal: 12,
    height: 38,
    borderRadius: 10,
    backgroundColor: colors.cardElevated,
    borderWidth: 1,
    borderColor: colors.border,
  },
  searchIcon: {
    marginRight: 8,
  },
  searchInput: {
    flex: 1,
    color: colors.text,
    fontSize: 14,
    paddingVertical: 0,
  },
  clearBtn: {
    padding: 4,
  },
  listContent: {
    paddingVertical: 4,
  },
  userRow: {
    flexDirection: "row",
    alignItems: "center",
    paddingHorizontal: 16,
    paddingVertical: 12,
    borderBottomWidth: 1,
    borderBottomColor: colors.line,
  },
  avatarWrap: {
    marginRight: 12,
  },
  avatarImage: {
    width: 46,
    height: 46,
    borderRadius: 23,
    backgroundColor: colors.cardElevated,
  },
  avatarPlaceholder: {
    width: 46,
    height: 46,
    borderRadius: 23,
    backgroundColor: colors.cardElevated,
    borderWidth: 1,
    borderColor: colors.border,
    alignItems: "center",
    justifyContent: "center",
  },
  avatarInitials: {
    color: colors.cyan,
    fontSize: 16,
    fontWeight: "700",
  },
  userInfo: {
    flex: 1,
    marginRight: 8,
  },
  nameRow: {
    flexDirection: "row",
    alignItems: "center",
    gap: 6,
  },
  displayName: {
    fontSize: 15,
    fontWeight: "700",
    color: colors.text,
  },
  mutualBadge: {
    paddingHorizontal: 5,
    paddingVertical: 1,
    borderRadius: 4,
    backgroundColor: "rgba(0, 229, 255, 0.15)",
    borderWidth: 1,
    borderColor: "rgba(0, 229, 255, 0.3)",
  },
  mutualBadgeText: {
    color: colors.cyan,
    fontSize: 10,
    fontWeight: "700",
  },
  username: {
    fontSize: 13,
    color: colors.textDim,
    marginTop: 1,
  },
  bioText: {
    fontSize: 12,
    color: colors.textMuted,
    marginTop: 3,
  },
  actions: {
    flexDirection: "row",
    alignItems: "center",
    gap: 8,
  },
  messageBtn: {
    width: 34,
    height: 32,
    borderRadius: 8,
    backgroundColor: colors.cardElevated,
    borderWidth: 1,
    borderColor: colors.border,
    alignItems: "center",
    justifyContent: "center",
  },
  followBtn: {
    paddingHorizontal: 14,
    height: 32,
    borderRadius: 8,
    alignItems: "center",
    justifyContent: "center",
  },
  notFollowingBtn: {
    backgroundColor: colors.cyan,
  },
  followingBtn: {
    backgroundColor: colors.cardElevated,
    borderWidth: 1,
    borderColor: colors.border,
  },
  followBtnText: {
    fontSize: 12,
    fontWeight: "700",
  },
  notFollowingBtnText: {
    color: "#000000",
  },
  followingBtnText: {
    color: colors.text,
  },
  loadingFooter: {
    paddingVertical: 16,
    alignItems: "center",
  },
  loadingContainer: {
    paddingVertical: 50,
    alignItems: "center",
  },
  emptyContainer: {
    alignItems: "center",
    justifyContent: "center",
    paddingHorizontal: 32,
    paddingVertical: 60,
    gap: 8,
  },
  emptyTitle: {
    fontSize: 16,
    fontWeight: "700",
    color: colors.text,
    textAlign: "center",
  },
  emptySubtitle: {
    fontSize: 13,
    color: colors.textDim,
    textAlign: "center",
    lineHeight: 18,
  },
});
