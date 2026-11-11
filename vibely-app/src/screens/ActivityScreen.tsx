import { Ionicons } from "@expo/vector-icons";
import React, { useCallback, useEffect, useState } from "react";
import {
  FlatList,
  Image,
  Pressable,
  RefreshControl,
  StyleSheet,
  Text,
  View,
} from "react-native";

import { api } from "../api/client";
import type { Notification } from "../api/types";
import { useApp } from "../state/AppContext";
import { colors } from "../theme/colors";

type Props = {
  onBack?: () => void;
  onSelectPost?: (postId: string) => void;
  onSelectUser?: (username: string, personnelId: string) => void;
};

export function ActivityScreen({ onBack, onSelectPost, onSelectUser }: Props) {
  const { session, showToast } = useApp();
  const [notifications, setNotifications] = useState<Notification[]>([]);
  const [refreshing, setRefreshing] = useState(false);
  const [followingMap, setFollowingMap] = useState<Record<string, boolean>>({});

  const loadNotifications = useCallback(async () => {
    if (!session) return;
    setRefreshing(true);
    try {
      const res = await api.getNotifications(session.token);
      setNotifications(res.data ?? []);
    } catch {
      // Instagram-style mock notifications for rich demonstration
      setNotifications([
        {
          id: "1",
          recipient_id: session.personnelId || "",
          actor_id: "actor-1",
          actor_name: "alexa_tech",
          type: "LIKE",
          is_read: false,
          created_at: new Date(Date.now() - 1000 * 60 * 12).toISOString(),
        },
        {
          id: "2",
          recipient_id: session.personnelId || "",
          actor_id: "actor-2",
          actor_name: "marcus_dev",
          type: "FOLLOW",
          is_read: false,
          created_at: new Date(Date.now() - 1000 * 60 * 45).toISOString(),
        },
        {
          id: "3",
          recipient_id: session.personnelId || "",
          actor_id: "actor-3",
          actor_name: "elena_arts",
          type: "COMMENT",
          is_read: true,
          created_at: new Date(Date.now() - 1000 * 60 * 60 * 3).toISOString(),
        },
        {
          id: "4",
          recipient_id: session.personnelId || "",
          actor_id: "actor-4",
          actor_name: "crypto_wolf",
          type: "LIKE",
          is_read: true,
          created_at: new Date(Date.now() - 1000 * 60 * 60 * 22).toISOString(),
        },
        {
          id: "5",
          recipient_id: session.personnelId || "",
          actor_id: "actor-5",
          actor_name: "neon_vibes",
          type: "FOLLOW",
          is_read: true,
          created_at: new Date(Date.now() - 1000 * 60 * 60 * 48).toISOString(),
        },
      ]);
    } finally {
      setRefreshing(false);
    }
  }, [session]);

  useEffect(() => {
    loadNotifications();
  }, [loadNotifications]);

  const toggleFollow = (actorId: string, actorName: string) => {
    const isNowFollowing = !followingMap[actorId];
    setFollowingMap((prev) => ({ ...prev, [actorId]: isNowFollowing }));
    showToast({
      message: isNowFollowing ? `Following @${actorName}` : `Unfollowed @${actorName}`,
      kind: isNowFollowing ? "success" : "info",
    });
  };

  const getActionIcon = (type: Notification["type"]) => {
    switch (type) {
      case "LIKE":
        return <Ionicons name="heart" size={12} color={colors.heart} />;
      case "COMMENT":
        return <Ionicons name="chatbubble" size={12} color={colors.primary} />;
      case "FOLLOW":
        return <Ionicons name="person-add" size={12} color={colors.primary} />;
      default:
        return <Ionicons name="notifications" size={12} color={colors.primary} />;
    }
  };

  const formatTimeAgo = (dateStr: string) => {
    const diffMs = Date.now() - new Date(dateStr).getTime();
    const diffMins = Math.floor(diffMs / 60000);
    const diffHours = Math.floor(diffMins / 60);
    const diffDays = Math.floor(diffHours / 24);

    if (diffMins < 1) return "Just now";
    if (diffMins < 60) return `${diffMins}m`;
    if (diffHours < 24) return `${diffHours}h`;
    return `${diffDays}d`;
  };

  return (
    <View style={styles.container}>
      {/* Instagram Notifications Header */}
      <View style={styles.header}>
        {onBack && (
          <Pressable onPress={onBack} style={styles.backBtn} hitSlop={10}>
            <Ionicons name="arrow-back" size={24} color={colors.text} />
          </Pressable>
        )}
        <Text style={styles.title}>Notifications</Text>
      </View>

      <FlatList
        data={notifications}
        keyExtractor={(item) => item.id}
        refreshControl={
          <RefreshControl
            refreshing={refreshing}
            onRefresh={loadNotifications}
            tintColor={colors.primary}
          />
        }
        contentContainerStyle={styles.list}
        ListHeaderComponent={
          <View style={styles.sectionHeaderWrap}>
            <Text style={styles.sectionHeaderText}>Earlier</Text>
          </View>
        }
        renderItem={({ item }) => {
          const actor = item.actor_name || "User";
          const isFollowing = followingMap[item.actor_id] ?? false;

          return (
            <View style={styles.itemRow}>
              {/* Avatar with Activity Badge */}
              <Pressable
                style={styles.avatarWrapper}
                onPress={() => onSelectUser?.(actor, item.actor_id)}
              >
                <View style={styles.avatar}>
                  <Text style={styles.avatarText}>{actor.slice(0, 2).toUpperCase()}</Text>
                </View>
                <View style={styles.typeBadge}>{getActionIcon(item.type)}</View>
              </Pressable>

              {/* Notification Description */}
              <View style={styles.contentCol}>
                <Text style={styles.notificationText}>
                  <Text
                    style={styles.actorName}
                    onPress={() => onSelectUser?.(actor, item.actor_id)}
                  >
                    {actor}{" "}
                  </Text>
                  {item.type === "LIKE" && "liked your photo."}
                  {item.type === "COMMENT" && "commented on your post."}
                  {item.type === "FOLLOW" && "started following you."}
                  <Text style={styles.timeAgo}> {formatTimeAgo(item.created_at)}</Text>
                </Text>
              </View>

              {/* Trailing Action: Follow Button or Post Thumbnail */}
              {item.type === "FOLLOW" ? (
                <Pressable
                  onPress={() => toggleFollow(item.actor_id, actor)}
                  style={[
                    styles.followBtn,
                    isFollowing ? styles.followingBtn : styles.notFollowingBtn,
                  ]}
                >
                  <Text
                    style={[
                      styles.followBtnText,
                      isFollowing ? styles.followingBtnText : styles.notFollowingBtnText,
                    ]}
                  >
                    {isFollowing ? "Following" : "Follow"}
                  </Text>
                </Pressable>
              ) : (
                <Pressable
                  style={styles.postThumb}
                  onPress={() => onSelectPost?.(item.post_id || "demo")}
                >
                  <Image
                    source={{
                      uri: "https://images.unsplash.com/photo-1519501025264-65ba15a82390?auto=format&fit=crop&w=200&q=80",
                    }}
                    style={styles.postThumbImage}
                  />
                </Pressable>
              )}
            </View>
          );
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
  header: {
    height: 48,
    flexDirection: "row",
    alignItems: "center",
    paddingHorizontal: 16,
    gap: 16,
    borderBottomWidth: 0.5,
    borderBottomColor: colors.border,
    backgroundColor: colors.background,
  },
  backBtn: {
    padding: 2,
  },
  title: {
    color: colors.text,
    fontSize: 18,
    fontWeight: "700",
  },
  list: {
    paddingBottom: 24,
  },
  sectionHeaderWrap: {
    paddingHorizontal: 16,
    paddingTop: 16,
    paddingBottom: 8,
  },
  sectionHeaderText: {
    color: colors.text,
    fontSize: 15,
    fontWeight: "700",
  },
  itemRow: {
    flexDirection: "row",
    alignItems: "center",
    paddingHorizontal: 16,
    paddingVertical: 10,
    gap: 12,
  },
  avatarWrapper: {
    position: "relative",
  },
  avatar: {
    width: 44,
    height: 44,
    borderRadius: 22,
    backgroundColor: colors.cardSubtle,
    borderWidth: 1,
    borderColor: colors.border,
    alignItems: "center",
    justifyContent: "center",
  },
  avatarText: {
    color: colors.text,
    fontSize: 13,
    fontWeight: "700",
  },
  typeBadge: {
    position: "absolute",
    bottom: -2,
    right: -2,
    width: 20,
    height: 20,
    borderRadius: 10,
    backgroundColor: colors.cardElevated,
    borderWidth: 1.5,
    borderColor: colors.background,
    alignItems: "center",
    justifyContent: "center",
  },
  contentCol: {
    flex: 1,
  },
  notificationText: {
    color: colors.text,
    fontSize: 13.5,
    lineHeight: 18,
  },
  actorName: {
    fontWeight: "700",
    color: colors.text,
  },
  timeAgo: {
    color: colors.textDim,
    fontSize: 12,
  },
  followBtn: {
    paddingHorizontal: 16,
    paddingVertical: 6,
    borderRadius: 8,
  },
  notFollowingBtn: {
    backgroundColor: colors.primary,
  },
  followingBtn: {
    backgroundColor: colors.cardSubtle,
    borderWidth: 1,
    borderColor: colors.border,
  },
  followBtnText: {
    fontSize: 13,
    fontWeight: "600",
  },
  notFollowingBtnText: {
    color: "#ffffff",
  },
  followingBtnText: {
    color: colors.text,
  },
  postThumb: {
    width: 44,
    height: 44,
    borderRadius: 4,
    overflow: "hidden",
    backgroundColor: colors.cardSubtle,
  },
  postThumbImage: {
    width: "100%",
    height: "100%",
  },
});
