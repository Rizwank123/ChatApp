import { Ionicons } from "@expo/vector-icons";
import React, { useCallback, useEffect, useState } from "react";
import {
  ActivityIndicator,
  Dimensions,
  FlatList,
  Image,
  Pressable,
  ScrollView,
  StyleSheet,
  Text,
  TextInput,
  View,
} from "react-native";

import { api } from "../api/client";
import type {
  HashtagSearchResult,
  Post,
  UserSearchResult,
} from "../api/types";
import { useApp } from "../state/AppContext";
import { colors } from "../theme/colors";

const { width } = Dimensions.get("window");
const GRID_GAP = 1.5;
const GRID_ITEM_SIZE = (width - GRID_GAP * 2) / 3;

const CATEGORIES = ["For you", "Trending", "Style", "Travel", "Art", "Tech", "Music", "Food"];
type SearchTab = "Top" | "Accounts" | "Tags" | "Posts";

type Props = {
  onSelectPost?: (post: Post) => void;
  onSelectUser?: (username: string, personnelId: string) => void;
};

export function ExploreScreen({ onSelectPost, onSelectUser }: Props) {
  const { session, showToast } = useApp();
  const [query, setQuery] = useState("");
  const [selectedCategory, setSelectedCategory] = useState("For you");
  const [activeTab, setActiveTab] = useState<SearchTab>("Top");
  const [loading, setLoading] = useState(false);

  // Search Results
  const [users, setUsers] = useState<UserSearchResult[]>([]);
  const [hashtags, setHashtags] = useState<HashtagSearchResult[]>([]);
  const [posts, setPosts] = useState<Post[]>([]);

  // Explore grid posts
  const [explorePosts, setExplorePosts] = useState<Post[]>([]);

  const loadExplore = useCallback(async () => {
    if (!session) return;
    try {
      const feedRes = await api.getExplorePosts(session.token, 0, 36);
      if (feedRes.data && feedRes.data.length > 0) {
        setExplorePosts(feedRes.data);
      } else {
        const fallback = await api.getFeed(session.token, 0, 36);
        setExplorePosts(fallback.data ?? []);
      }
    } catch {
      try {
        const fallback = await api.getFeed(session.token, 0, 36);
        setExplorePosts(fallback.data ?? []);
      } catch {
        // quiet fallback
      }
    }
  }, [session]);

  const runSearch = useCallback(async () => {
    if (!session || !query.trim()) {
      setUsers([]);
      setHashtags([]);
      setPosts([]);
      return;
    }

    setLoading(true);
    try {
      const cleanQuery = query.trim();
      if (cleanQuery.startsWith("#")) {
        const tag = cleanQuery.replace(/^#/, "");
        const [tagDetails, tagPosts] = await Promise.all([
          api.searchHashtags(session.token, tag, 0, 10).catch(() => ({ data: [] })),
          api.getHashtagPosts(session.token, tag, 0, 30).catch(() => ({ data: [] })),
        ]);
        setHashtags(tagDetails.data ?? []);
        setPosts(tagPosts.data ?? []);
        setUsers([]);
      } else {
        const unified = await api.unifiedSearch(session.token, cleanQuery, 15);
        if (unified.data) {
          setUsers(unified.data.users ?? []);
          setHashtags(unified.data.hashtags ?? []);
          setPosts(unified.data.posts ?? []);
        }
      }
    } catch {
      // search silent error fallback
    } finally {
      setLoading(false);
    }
  }, [query, session]);

  useEffect(() => {
    loadExplore();
  }, [loadExplore]);

  useEffect(() => {
    const t = setTimeout(() => {
      runSearch();
    }, 280);
    return () => clearTimeout(t);
  }, [runSearch]);

  const toggleFollow = async (person: UserSearchResult) => {
    if (!session) return;
    const wasFollowing = person.is_following;
    const nextFollowing = !wasFollowing;

    setUsers((prev) =>
      prev.map((p) =>
        p.personnel_id === person.personnel_id
          ? { ...p, is_following: nextFollowing }
          : p,
      ),
    );

    try {
      if (wasFollowing) {
        await api.unfollowSocial(session.token, person.personnel_id);
        showToast({ message: `Unfollowed @${person.user_name}`, kind: "info" });
      } else {
        await api.followSocial(session.token, person.personnel_id);
        showToast({ message: `Following @${person.user_name}`, kind: "success" });
      }
    } catch {
      setUsers((prev) =>
        prev.map((p) =>
          p.personnel_id === person.personnel_id
            ? { ...p, is_following: wasFollowing }
            : p,
        ),
      );
      showToast({ message: "Action failed", kind: "error" });
    }
  };

  const isSearching = query.trim().length > 0;

  return (
    <View style={styles.container}>
      {/* Instagram Search Header */}
      <View style={styles.header}>
        <View style={styles.searchBar}>
          <Ionicons name="search" size={18} color={colors.textDim} />
          <TextInput
            placeholder="Search"
            placeholderTextColor={colors.textDim}
            style={styles.searchInput}
            value={query}
            onChangeText={setQuery}
            autoCapitalize="none"
          />
          {query.length > 0 && (
            <Pressable onPress={() => setQuery("")} hitSlop={8}>
              <Ionicons name="close-circle" size={16} color={colors.textDim} />
            </Pressable>
          )}
        </View>
      </View>

      {/* Tabs when search is active */}
      {isSearching && (
        <View style={styles.searchTabs}>
          {(["Top", "Accounts", "Tags", "Posts"] as SearchTab[]).map((tab) => (
            <Pressable
              key={tab}
              onPress={() => setActiveTab(tab)}
              style={[styles.searchTabItem, activeTab === tab && styles.searchTabActive]}
            >
              <Text
                style={[
                  styles.searchTabText,
                  activeTab === tab && styles.searchTabActiveText,
                ]}
              >
                {tab}
              </Text>
            </Pressable>
          ))}
        </View>
      )}

      {loading && (
        <View style={styles.loadingBar}>
          <ActivityIndicator size="small" color={colors.primary} />
        </View>
      )}

      {/* Active Search Mode */}
      {isSearching ? (
        <ScrollView contentContainerStyle={styles.scrollContent}>
          {/* Top Tab: Mixed Preview */}
          {activeTab === "Top" && (
            <View>
              {users.length > 0 && (
                <View style={styles.section}>
                  <Text style={styles.sectionHeader}>Accounts</Text>
                  {users.slice(0, 4).map((u) =>
                    renderUserRow(u, toggleFollow, onSelectUser),
                  )}
                </View>
              )}

              {hashtags.length > 0 && (
                <View style={styles.section}>
                  <Text style={styles.sectionHeader}>Hashtags</Text>
                  {hashtags
                    .slice(0, 4)
                    .map((h) => renderHashtagRow(h, (tag) => setQuery(`#${tag}`)))}
                </View>
              )}

              {posts.length > 0 && (
                <View style={styles.section}>
                  <Text style={styles.sectionHeader}>Posts</Text>
                  {renderPostsGrid(posts, onSelectPost)}
                </View>
              )}

              {users.length === 0 && hashtags.length === 0 && posts.length === 0 && !loading && (
                <Text style={styles.emptySearchText}>No results found for "{query}"</Text>
              )}
            </View>
          )}

          {/* Accounts Tab */}
          {activeTab === "Accounts" && (
            <View style={styles.section}>
              {users.length > 0 ? (
                users.map((u) => renderUserRow(u, toggleFollow, onSelectUser))
              ) : (
                !loading && <Text style={styles.emptySearchText}>No accounts found</Text>
              )}
            </View>
          )}

          {/* Tags Tab */}
          {activeTab === "Tags" && (
            <View style={styles.section}>
              {hashtags.length > 0 ? (
                hashtags.map((h) => renderHashtagRow(h, (tag) => setQuery(`#${tag}`)))
              ) : (
                !loading && <Text style={styles.emptySearchText}>No hashtags found</Text>
              )}
            </View>
          )}

          {/* Posts Tab */}
          {activeTab === "Posts" && (
            <View style={styles.section}>
              {posts.length > 0 ? (
                renderPostsGrid(posts, onSelectPost)
              ) : (
                !loading && <Text style={styles.emptySearchText}>No posts found</Text>
              )}
            </View>
          )}
        </ScrollView>
      ) : (
        /* Explore / Browse Mode */
        <>
          {/* Category Chips */}
          <View style={styles.categoriesWrapper}>
            <ScrollView
              horizontal
              showsHorizontalScrollIndicator={false}
              contentContainerStyle={styles.categoryScroll}
            >
              {CATEGORIES.map((cat) => {
                const isSelected = selectedCategory === cat;
                return (
                  <Pressable
                    key={cat}
                    onPress={() => setSelectedCategory(cat)}
                    style={[styles.pill, isSelected && styles.activePill]}
                  >
                    <Text style={[styles.pillText, isSelected && styles.activePillText]}>
                      {cat}
                    </Text>
                  </Pressable>
                );
              })}
            </ScrollView>
          </View>

          {/* 3-Column Instagram Photo Grid */}
          <FlatList
            data={explorePosts}
            keyExtractor={(item) => item.id}
            numColumns={3}
            contentContainerStyle={styles.gridContainer}
            renderItem={({ item }) => {
              const media = item.media_url?.[0] || item.media?.[0]?.url;
              const hasMultiple =
                (item.media_url && item.media_url.length > 1) ||
                (item.media && item.media.length > 1);

              return (
                <Pressable
                  onPress={() => onSelectPost && onSelectPost(item)}
                  style={styles.gridItem}
                >
                  {media ? (
                    <Image source={{ uri: media }} style={styles.gridImage} resizeMode="cover" />
                  ) : (
                    <View style={styles.gridPlaceholder}>
                      <Ionicons name="sparkles" size={20} color={colors.textDim} />
                      <Text style={styles.gridCaption} numberOfLines={2}>
                        {item.caption || "Vibely"}
                      </Text>
                    </View>
                  )}

                  {/* Instagram Multi-Image corner icon */}
                  {hasMultiple && (
                    <View style={styles.multiIconBadge}>
                      <Ionicons name="copy-outline" size={13} color="#ffffff" />
                    </View>
                  )}
                </Pressable>
              );
            }}
            ListEmptyComponent={
              <View style={styles.emptyGrid}>
                <Ionicons name="compass-outline" size={54} color={colors.textDim} />
                <Text style={styles.emptyGridText}>No explore items yet</Text>
              </View>
            }
          />
        </>
      )}
    </View>
  );
}

function renderUserRow(
  item: UserSearchResult,
  toggleFollow: (person: UserSearchResult) => void,
  onSelectUser?: (username: string, personnelId: string) => void,
) {
  const isFollowing = item.is_following;
  const followsYou = item.follows_you;

  let buttonLabel = "Follow";
  if (isFollowing) {
    buttonLabel = "Following";
  } else if (followsYou) {
    buttonLabel = "Follow back";
  }

  return (
    <View key={item.personnel_id} style={styles.personRow}>
      <Pressable
        style={styles.personRowLeft}
        onPress={() => onSelectUser?.(item.user_name, item.personnel_id)}
      >
        <View style={styles.avatar}>
          <Text style={styles.avatarInitial}>
            {(item.display_name || item.user_name).slice(0, 2).toUpperCase()}
          </Text>
        </View>
        <View style={styles.personMeta}>
          <Text style={styles.personName}>{item.user_name}</Text>
          <Text style={styles.personHandle}>{item.display_name || item.user_name}</Text>
          {item.bio ? (
            <Text style={styles.personBio} numberOfLines={1}>
              {item.bio}
            </Text>
          ) : null}
        </View>
      </Pressable>

      <Pressable
        onPress={() => toggleFollow(item)}
        style={[
          styles.followBtn,
          isFollowing && styles.followingBtn,
          !isFollowing && styles.notFollowingBtn,
        ]}
      >
        <Text
          style={[
            styles.followBtnText,
            isFollowing && styles.followingBtnText,
            !isFollowing && styles.notFollowingBtnText,
          ]}
        >
          {buttonLabel}
        </Text>
      </Pressable>
    </View>
  );
}

function renderHashtagRow(
  item: HashtagSearchResult,
  onPress: (tag: string) => void,
) {
  return (
    <Pressable
      key={item.id || item.name}
      style={styles.hashtagRow}
      onPress={() => onPress(item.name)}
    >
      <View style={styles.hashtagCircle}>
        <Ionicons name="pricetag-outline" size={18} color={colors.text} />
      </View>
      <View style={styles.hashtagMeta}>
        <Text style={styles.hashtagName}>#{item.name}</Text>
        <Text style={styles.hashtagCount}>
          {item.post_count ? `${item.post_count.toLocaleString()} posts` : "Popular hashtag"}
        </Text>
      </View>
    </Pressable>
  );
}

function renderPostsGrid(posts: Post[], onSelectPost?: (post: Post) => void) {
  return (
    <View style={styles.searchGridContainer}>
      {posts.map((post) => {
        const media = post.media_url?.[0] || post.media?.[0]?.url;
        return (
          <Pressable
            key={post.id}
            onPress={() => onSelectPost?.(post)}
            style={styles.searchGridItem}
          >
            {media ? (
              <Image source={{ uri: media }} style={styles.gridImage} resizeMode="cover" />
            ) : (
              <View style={styles.gridPlaceholder}>
                <Ionicons name="image-outline" size={20} color={colors.textDim} />
              </View>
            )}
          </Pressable>
        );
      })}
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: colors.background,
  },
  header: {
    paddingHorizontal: 12,
    paddingVertical: 8,
    backgroundColor: colors.background,
  },
  searchBar: {
    height: 38,
    borderRadius: 10,
    backgroundColor: colors.cardSubtle,
    flexDirection: "row",
    alignItems: "center",
    paddingHorizontal: 12,
    gap: 8,
  },
  searchInput: {
    flex: 1,
    color: colors.text,
    fontSize: 14,
  },
  categoriesWrapper: {
    paddingVertical: 6,
    borderBottomWidth: 0.5,
    borderBottomColor: colors.border,
  },
  categoryScroll: {
    paddingHorizontal: 12,
    gap: 8,
  },
  pill: {
    paddingHorizontal: 14,
    paddingVertical: 6,
    borderRadius: 8,
    backgroundColor: colors.cardSubtle,
    borderWidth: 1,
    borderColor: colors.border,
  },
  activePill: {
    backgroundColor: colors.cardElevated,
    borderColor: colors.textMuted,
  },
  pillText: {
    color: colors.textMuted,
    fontSize: 13,
    fontWeight: "600",
  },
  activePillText: {
    color: colors.text,
  },
  gridContainer: {
    paddingBottom: 24,
  },
  gridItem: {
    width: GRID_ITEM_SIZE,
    height: GRID_ITEM_SIZE,
    marginRight: GRID_GAP,
    marginBottom: GRID_GAP,
    backgroundColor: colors.cardSubtle,
    position: "relative",
  },
  gridImage: {
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
  multiIconBadge: {
    position: "absolute",
    top: 6,
    right: 6,
  },
  emptyGrid: {
    padding: 40,
    alignItems: "center",
    justifyContent: "center",
    gap: 12,
    marginTop: 60,
  },
  emptyGridText: {
    color: colors.textDim,
    fontSize: 14,
  },
  searchTabs: {
    flexDirection: "row",
    borderBottomWidth: 0.5,
    borderBottomColor: colors.border,
    backgroundColor: colors.background,
  },
  searchTabItem: {
    flex: 1,
    paddingVertical: 10,
    alignItems: "center",
  },
  searchTabActive: {
    borderBottomWidth: 1.5,
    borderBottomColor: colors.text,
  },
  searchTabText: {
    color: colors.textDim,
    fontSize: 13.5,
    fontWeight: "600",
  },
  searchTabActiveText: {
    color: colors.text,
  },
  loadingBar: {
    paddingVertical: 8,
    alignItems: "center",
  },
  scrollContent: {
    paddingBottom: 24,
  },
  section: {
    paddingTop: 10,
  },
  sectionHeader: {
    color: colors.textDim,
    fontSize: 12,
    fontWeight: "700",
    textTransform: "uppercase",
    letterSpacing: 0.5,
    paddingHorizontal: 16,
    marginBottom: 6,
  },
  personRow: {
    flexDirection: "row",
    alignItems: "center",
    justifyContent: "space-between",
    paddingHorizontal: 16,
    paddingVertical: 8,
  },
  personRowLeft: {
    flexDirection: "row",
    alignItems: "center",
    gap: 12,
    flex: 1,
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
  avatarInitial: {
    color: colors.text,
    fontSize: 14,
    fontWeight: "700",
  },
  personMeta: {
    flex: 1,
  },
  personName: {
    color: colors.text,
    fontSize: 13.5,
    fontWeight: "600",
  },
  personHandle: {
    color: colors.textDim,
    fontSize: 12.5,
  },
  personBio: {
    color: colors.textMuted,
    fontSize: 11.5,
    marginTop: 1,
  },
  followBtn: {
    paddingHorizontal: 16,
    paddingVertical: 6,
    borderRadius: 8,
    alignItems: "center",
    justifyContent: "center",
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
  hashtagRow: {
    flexDirection: "row",
    alignItems: "center",
    paddingHorizontal: 16,
    paddingVertical: 10,
    gap: 14,
  },
  hashtagCircle: {
    width: 44,
    height: 44,
    borderRadius: 22,
    borderWidth: 1,
    borderColor: colors.border,
    alignItems: "center",
    justifyContent: "center",
    backgroundColor: colors.cardSubtle,
  },
  hashtagMeta: {
    flex: 1,
  },
  hashtagName: {
    color: colors.text,
    fontSize: 13.5,
    fontWeight: "600",
  },
  hashtagCount: {
    color: colors.textDim,
    fontSize: 12,
  },
  searchGridContainer: {
    flexDirection: "row",
    flexWrap: "wrap",
  },
  searchGridItem: {
    width: GRID_ITEM_SIZE,
    height: GRID_ITEM_SIZE,
    marginRight: GRID_GAP,
    marginBottom: GRID_GAP,
    backgroundColor: colors.cardSubtle,
  },
  emptySearchText: {
    color: colors.textDim,
    textAlign: "center",
    padding: 32,
    fontSize: 13.5,
  },
});
