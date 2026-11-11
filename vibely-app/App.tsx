import { Ionicons } from "@expo/vector-icons";
import { StatusBar } from "expo-status-bar";
import React, { useState } from "react";
import {
  ActivityIndicator,
  Modal,
  Pressable,
  SafeAreaView,
  StyleSheet,
  Text,
  View,
} from "react-native";

import type { Post, UUID } from "./src/api/types";
import { Toast } from "./src/components/Toast";
import { ActivityScreen } from "./src/screens/ActivityScreen";
import { AuthScreen } from "./src/screens/AuthScreen";
import { ChatScreen } from "./src/screens/ChatScreen";
import { CommentsScreen } from "./src/screens/CommentsScreen";
import { ComposerScreen } from "./src/screens/ComposerScreen";
import { ExploreScreen } from "./src/screens/ExploreScreen";
import { FeedScreen } from "./src/screens/FeedScreen";
import { FollowListScreen } from "./src/screens/FollowListScreen";
import { ProfileScreen } from "./src/screens/ProfileScreen";
import { AppProvider, useApp } from "./src/state/AppContext";
import { colors } from "./src/theme/colors";

type Tab = "feed" | "explore" | "create" | "chat" | "profile";

function AppContent() {
  const { session, setSession, isRestoringSession } = useApp();
  const [tab, setTab] = useState<Tab>("feed");
  const [commentsPost, setCommentsPost] = useState<Post>();
  const [showActivity, setShowActivity] = useState(false);
  const [followListParams, setFollowListParams] = useState<{
    initialTab: "followers" | "following";
    followerCount: number;
    followingCount: number;
  } | null>(null);
  const [directChatTarget, setDirectChatTarget] = useState<{
    participantId: UUID;
    participantName: string;
  } | null>(null);

  if (isRestoringSession) {
    return (
      <SafeAreaView style={[styles.safe, styles.center]}>
        <ActivityIndicator size="large" color={colors.primary} />
        <StatusBar style="light" />
      </SafeAreaView>
    );
  }

  if (!session) {
    return (
      <SafeAreaView style={styles.safe}>
        <AuthScreen />
        <Toast />
        <StatusBar style="light" />
      </SafeAreaView>
    );
  }

  if (commentsPost) {
    return (
      <SafeAreaView style={styles.safe}>
        <CommentsScreen post={commentsPost} onBack={() => setCommentsPost(undefined)} />
        <Toast />
        <StatusBar style="light" />
      </SafeAreaView>
    );
  }

  const currentUserInitial = (session.userName || "U").slice(0, 2).toUpperCase();

  return (
    <SafeAreaView style={styles.safe}>
      {/* Instagram App Header (Shown on Feed) */}
      {tab === "feed" && (
        <View style={styles.instagramHeader}>
          <View style={styles.logoRow}>
            <Text style={styles.logoText}>Vibely</Text>
            <Ionicons name="chevron-down" size={16} color={colors.text} style={{ marginTop: 2 }} />
          </View>

          <View style={styles.headerIcons}>
            <Pressable
              onPress={() => setShowActivity(true)}
              style={styles.headerIconButton}
              hitSlop={8}
            >
              <Ionicons name="heart-outline" size={25} color={colors.text} />
              <View style={styles.badgeDot} />
            </Pressable>

            <Pressable
              onPress={() => setTab("chat")}
              style={styles.headerIconButton}
              hitSlop={8}
            >
              <Ionicons name="paper-plane-outline" size={24} color={colors.text} />
            </Pressable>
          </View>
        </View>
      )}

      {/* Screen Router */}
      <View style={styles.main}>
        {tab === "feed" ? (
          <FeedScreen
            onOpenComments={setCommentsPost}
            onOpenCreateStory={() => setTab("create")}
            onSelectUser={(uName, pId) => {
              setDirectChatTarget({ participantId: pId, participantName: uName });
            }}
          />
        ) : null}

        {tab === "explore" ? (
          <ExploreScreen
            onSelectPost={setCommentsPost}
            onSelectUser={(uName, pId) => {
              setDirectChatTarget({ participantId: pId, participantName: uName });
            }}
          />
        ) : null}

        {tab === "create" ? <ComposerScreen onCreated={() => setTab("feed")} /> : null}

        {tab === "chat" ? <ChatScreen /> : null}

        {tab === "profile" ? (
          <ProfileScreen
            onSelectPost={setCommentsPost}
            onLogout={() => setSession(undefined)}
            onOpenCreate={() => setTab("create")}
            onOpenFollowList={(initialTab, followerCount, followingCount) =>
              setFollowListParams({ initialTab, followerCount, followingCount })
            }
          />
        ) : null}
      </View>

      {/* Followers / Following List Modal */}
      <Modal visible={!!followListParams} animationType="slide">
        <SafeAreaView style={styles.safe}>
          {followListParams && session?.personnelId && (
            <View style={{ flex: 1, display: directChatTarget ? "none" : "flex" }}>
              <FollowListScreen
                targetPersonnelId={session.personnelId}
                targetUsername={session.userName || "Your Profile"}
                initialTab={followListParams.initialTab}
                followerCount={followListParams.followerCount}
                followingCount={followListParams.followingCount}
                onBack={() => setFollowListParams(null)}
                onOpenChat={(participantId, participantName) =>
                  setDirectChatTarget({ participantId, participantName })
                }
              />
            </View>
          )}
          {directChatTarget && (
            <View style={{ flex: 1 }}>
              <ChatScreen
                initialParticipantId={directChatTarget.participantId}
                initialParticipantName={directChatTarget.participantName}
                onBack={() => setDirectChatTarget(null)}
              />
            </View>
          )}
          <Toast />
          <StatusBar style="light" />
        </SafeAreaView>
      </Modal>

      {/* Activity / Notifications Modal */}
      <Modal visible={showActivity} animationType="slide">
        <SafeAreaView style={styles.safe}>
          <ActivityScreen
            onBack={() => setShowActivity(false)}
            onSelectUser={(uName, pId) => {
              setShowActivity(false);
              setDirectChatTarget({ participantId: pId, participantName: uName });
            }}
          />
          <Toast />
          <StatusBar style="light" />
        </SafeAreaView>
      </Modal>

      {/* 5-Tab Instagram Bottom Navigation Bar */}
      <View style={styles.tabs}>
        <TabButton
          active={tab === "feed"}
          icon={tab === "feed" ? "home" : "home-outline"}
          onPress={() => setTab("feed")}
        />
        <TabButton
          active={tab === "explore"}
          icon={tab === "explore" ? "search" : "search-outline"}
          onPress={() => setTab("explore")}
        />
        <TabButton
          active={tab === "create"}
          icon={tab === "create" ? "add-circle" : "add-circle-outline"}
          onPress={() => setTab("create")}
        />
        <TabButton
          active={tab === "chat"}
          icon={tab === "chat" ? "paper-plane" : "paper-plane-outline"}
          onPress={() => setTab("chat")}
        />

        {/* Profile Avatar Tab (Instagram Style) */}
        <Pressable onPress={() => setTab("profile")} style={styles.tabButton}>
          <View
            style={[
              styles.profileTabAvatarWrap,
              tab === "profile" && styles.profileTabAvatarActive,
            ]}
          >
            <View style={styles.profileTabAvatarInner}>
              <Text style={styles.profileTabAvatarText}>{currentUserInitial}</Text>
            </View>
          </View>
        </Pressable>
      </View>

      <Toast />
      <StatusBar style="light" />
    </SafeAreaView>
  );
}

function TabButton({
  active,
  icon,
  onPress,
}: {
  active: boolean;
  icon: keyof typeof Ionicons.glyphMap;
  onPress: () => void;
}) {
  return (
    <Pressable onPress={onPress} style={styles.tabButton} hitSlop={6}>
      <Ionicons
        name={icon}
        size={26}
        color={active ? colors.text : colors.textMuted}
      />
    </Pressable>
  );
}

export default function App() {
  return (
    <AppProvider>
      <AppContent />
    </AppProvider>
  );
}

const styles = StyleSheet.create({
  safe: {
    flex: 1,
    backgroundColor: colors.background,
  },
  center: {
    justifyContent: "center",
    alignItems: "center",
  },
  instagramHeader: {
    height: 50,
    flexDirection: "row",
    alignItems: "center",
    justifyContent: "space-between",
    paddingHorizontal: 16,
    backgroundColor: colors.background,
    borderBottomWidth: 0.5,
    borderBottomColor: colors.border,
  },
  logoRow: {
    flexDirection: "row",
    alignItems: "center",
    gap: 6,
  },
  logoText: {
    color: colors.text,
    fontSize: 25,
    fontWeight: "900",
    letterSpacing: -0.7,
  },
  headerIcons: {
    flexDirection: "row",
    alignItems: "center",
    gap: 18,
  },
  headerIconButton: {
    position: "relative",
    padding: 2,
  },
  badgeDot: {
    position: "absolute",
    top: 2,
    right: 2,
    width: 8,
    height: 8,
    borderRadius: 4,
    backgroundColor: colors.heart,
  },
  main: {
    flex: 1,
    backgroundColor: colors.background,
  },
  tabs: {
    height: 52,
    flexDirection: "row",
    alignItems: "center",
    justifyContent: "space-around",
    backgroundColor: colors.background,
    borderTopWidth: 0.5,
    borderTopColor: colors.border,
    paddingBottom: 2,
  },
  tabButton: {
    flex: 1,
    height: "100%",
    alignItems: "center",
    justifyContent: "center",
  },
  profileTabAvatarWrap: {
    width: 28,
    height: 28,
    borderRadius: 14,
    borderWidth: 1.5,
    borderColor: "transparent",
    alignItems: "center",
    justifyContent: "center",
  },
  profileTabAvatarActive: {
    borderColor: colors.text,
  },
  profileTabAvatarInner: {
    width: 22,
    height: 22,
    borderRadius: 11,
    backgroundColor: colors.cardSubtle,
    borderWidth: 1,
    borderColor: colors.border,
    alignItems: "center",
    justifyContent: "center",
  },
  profileTabAvatarText: {
    color: colors.text,
    fontSize: 9.5,
    fontWeight: "700",
  },
});
