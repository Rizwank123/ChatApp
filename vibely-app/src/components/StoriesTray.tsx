import { Ionicons } from "@expo/vector-icons";
import React from "react";
import { FlatList, Pressable, StyleSheet, Text, View } from "react-native";

import { colors } from "../theme/colors";

export type StoryItem = {
  id: string;
  username: string;
  isUser?: boolean;
  hasUnseen?: boolean;
  avatarUrl?: string;
};

type Props = {
  stories: StoryItem[];
  onSelectStory?: (story: StoryItem, index: number) => void;
  onAddStory?: () => void;
};

export function StoriesTray({ stories, onSelectStory, onAddStory }: Props) {
  return (
    <View style={styles.container}>
      <FlatList
        horizontal
        showsHorizontalScrollIndicator={false}
        contentContainerStyle={styles.list}
        data={stories}
        keyExtractor={(item) => item.id}
        renderItem={({ item, index }) => {
          if (item.isUser) {
            return (
              <Pressable style={styles.storyCol} onPress={onAddStory}>
                <View style={styles.userAvatarContainer}>
                  <View style={styles.avatarInner}>
                    <Ionicons name="person" size={26} color={colors.textMuted} />
                  </View>
                  <View style={styles.addBadge}>
                    <Ionicons name="add" size={14} color="#ffffff" />
                  </View>
                </View>
                <Text style={styles.storyName} numberOfLines={1}>
                  Your story
                </Text>
              </Pressable>
            );
          }

          return (
            <Pressable
              style={styles.storyCol}
              onPress={() => onSelectStory && onSelectStory(item, index)}
            >
              {/* Instagram Story Ring: Outer gradient effect & inner cutout */}
              <View
                style={[
                  styles.storyRingOuter,
                  item.hasUnseen ? styles.activeRingOuter : styles.seenRingOuter,
                ]}
              >
                <View style={styles.storyRingGap}>
                  <View style={styles.avatarInner}>
                    <Text style={styles.avatarInitial}>
                      {item.username.slice(0, 2).toUpperCase()}
                    </Text>
                  </View>
                </View>
              </View>
              <Text style={styles.storyName} numberOfLines={1}>
                {item.username}
              </Text>
            </Pressable>
          );
        }}
      />
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    paddingVertical: 10,
    borderBottomWidth: 0.5,
    borderBottomColor: colors.border,
    backgroundColor: colors.canvas,
  },
  list: {
    paddingHorizontal: 12,
    gap: 16,
  },
  storyCol: {
    alignItems: "center",
    width: 68,
    gap: 4,
  },
  userAvatarContainer: {
    width: 68,
    height: 68,
    borderRadius: 34,
    borderWidth: 1,
    borderColor: colors.border,
    alignItems: "center",
    justifyContent: "center",
    position: "relative",
  },
  storyRingOuter: {
    width: 68,
    height: 68,
    borderRadius: 34,
    padding: 2.5,
    alignItems: "center",
    justifyContent: "center",
  },
  activeRingOuter: {
    backgroundColor: colors.storyPink,
    borderWidth: 2,
    borderColor: colors.storyOrange,
  },
  seenRingOuter: {
    backgroundColor: colors.storySeen,
    borderWidth: 1,
    borderColor: colors.border,
  },
  storyRingGap: {
    width: 60,
    height: 60,
    borderRadius: 30,
    backgroundColor: colors.canvas,
    alignItems: "center",
    justifyContent: "center",
    padding: 2,
  },
  avatarInner: {
    width: 56,
    height: 56,
    borderRadius: 28,
    backgroundColor: colors.cardSubtle,
    alignItems: "center",
    justifyContent: "center",
  },
  avatarInitial: {
    color: colors.text,
    fontSize: 15,
    fontWeight: "700",
  },
  addBadge: {
    position: "absolute",
    bottom: 1,
    right: 1,
    width: 22,
    height: 22,
    borderRadius: 11,
    backgroundColor: colors.primary,
    borderWidth: 2.5,
    borderColor: colors.canvas,
    alignItems: "center",
    justifyContent: "center",
  },
  storyName: {
    color: colors.text,
    fontSize: 11.5,
    fontWeight: "400",
    textAlign: "center",
    maxWidth: 68,
  },
});
