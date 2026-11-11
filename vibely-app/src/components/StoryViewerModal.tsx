import { Ionicons } from "@expo/vector-icons";
import React, { useEffect, useRef, useState } from "react";
import {
  Animated,
  Dimensions,
  Image,
  Modal,
  Pressable,
  SafeAreaView,
  StyleSheet,
  Text,
  TextInput,
  View,
} from "react-native";

import { colors } from "../theme/colors";
import type { StoryItem } from "./StoriesTray";

const { width, height } = Dimensions.get("window");

type Props = {
  visible: boolean;
  stories: StoryItem[];
  initialIndex?: number;
  onClose: () => void;
  onReply?: (story: StoryItem, message: string) => void;
  onLikeStory?: (story: StoryItem) => void;
};

// Curated high quality story images for mock/demo stories
const STORY_IMAGES: Record<string, string> = {
  alexa: "https://images.unsplash.com/photo-1534528741775-53994a69daeb?auto=format&fit=crop&w=1080&q=80",
  marcus_dev: "https://images.unsplash.com/photo-1507003211169-0a1dd7228f2d?auto=format&fit=crop&w=1080&q=80",
  neon_vibes: "https://images.unsplash.com/photo-1518770660439-4636190af475?auto=format&fit=crop&w=1080&q=80",
  elena_arts: "https://images.unsplash.com/photo-1531746020798-e6953c6e8e04?auto=format&fit=crop&w=1080&q=80",
  crypto_wolf: "https://images.unsplash.com/photo-1517841905240-472988babdf9?auto=format&fit=crop&w=1080&q=80",
  You: "https://images.unsplash.com/photo-1522075469751-3a6694fb2f61?auto=format&fit=crop&w=1080&q=80",
};

export function StoryViewerModal({
  visible,
  stories,
  initialIndex = 0,
  onClose,
  onReply,
  onLikeStory,
}: Props) {
  const [currentIndex, setCurrentIndex] = useState(initialIndex);
  const [replyText, setReplyText] = useState("");
  const [isLiked, setIsLiked] = useState(false);
  const [isPaused, setIsPaused] = useState(false);

  const progressAnim = useRef(new Animated.Value(0)).current;
  const animRef = useRef<Animated.CompositeAnimation | null>(null);

  const activeStories = stories.filter((s) => !s.isUser || s.hasUnseen);
  const currentStory = activeStories[currentIndex] || stories[0];

  useEffect(() => {
    if (visible) {
      setCurrentIndex(Math.min(initialIndex, activeStories.length - 1));
      setIsLiked(false);
      setReplyText("");
    }
  }, [visible, initialIndex, activeStories.length]);

  useEffect(() => {
    if (!visible || !currentStory) return;

    progressAnim.setValue(0);
    setIsLiked(false);

    if (isPaused) return;

    animRef.current = Animated.timing(progressAnim, {
      toValue: 1,
      duration: 5000,
      useNativeDriver: false,
    });

    animRef.current.start(({ finished }) => {
      if (finished) {
        handleNext();
      }
    });

    return () => {
      animRef.current?.stop();
    };
  }, [currentIndex, visible, isPaused]);

  const handleNext = () => {
    if (currentIndex < activeStories.length - 1) {
      setCurrentIndex((prev) => prev + 1);
    } else {
      onClose();
    }
  };

  const handlePrev = () => {
    if (currentIndex > 0) {
      setCurrentIndex((prev) => prev - 1);
    } else {
      progressAnim.setValue(0);
    }
  };

  const handleTap = (evt: any) => {
    const touchX = evt.nativeEvent.locationX;
    if (touchX < width * 0.3) {
      handlePrev();
    } else {
      handleNext();
    }
  };

  const handleSendReply = () => {
    if (replyText.trim() && currentStory) {
      onReply?.(currentStory, replyText.trim());
      setReplyText("");
    }
  };

  const handleToggleLike = () => {
    setIsLiked((prev) => !prev);
    if (currentStory && !isLiked) {
      onLikeStory?.(currentStory);
    }
  };

  if (!visible || !currentStory) return null;

  const imageUrl =
    STORY_IMAGES[currentStory.username] ||
    "https://images.unsplash.com/photo-1579783900882-c0d3dad7b119?auto=format&fit=crop&w=1080&q=80";

  return (
    <Modal visible={visible} animationType="fade" transparent onRequestClose={onClose}>
      <SafeAreaView style={styles.container}>
        {/* Top Progress Bars */}
        <View style={styles.progressContainer}>
          {activeStories.map((s, idx) => {
            let fillWidth: any = "0%";
            if (idx < currentIndex) {
              fillWidth = "100%";
            } else if (idx === currentIndex) {
              fillWidth = progressAnim.interpolate({
                inputRange: [0, 1],
                outputRange: ["0%", "100%"],
              });
            }
            return (
              <View key={s.id || idx} style={styles.progressBarTrack}>
                <Animated.View style={[styles.progressBarFill, { width: fillWidth }]} />
              </View>
            );
          })}
        </View>

        {/* Story Header */}
        <View style={styles.header}>
          <View style={styles.userInfo}>
            <View style={styles.avatarBorder}>
              <View style={styles.avatar}>
                <Text style={styles.avatarText}>
                  {currentStory.username.slice(0, 2).toUpperCase()}
                </Text>
              </View>
            </View>
            <Text style={styles.username}>{currentStory.username}</Text>
            <Text style={styles.timeAgo}>2h</Text>
          </View>

          <Pressable onPress={onClose} style={styles.closeBtn}>
            <Ionicons name="close" size={26} color="#ffffff" />
          </Pressable>
        </View>

        {/* Story Media & Tap Zones */}
        <Pressable
          style={styles.mediaContainer}
          onPress={handleTap}
          onPressIn={() => setIsPaused(true)}
          onPressOut={() => setIsPaused(false)}
        >
          <Image source={{ uri: imageUrl }} style={styles.storyImage} resizeMode="cover" />
        </Pressable>

        {/* Bottom Reply Bar */}
        <View style={styles.bottomBar}>
          <View style={styles.replyInputWrap}>
            <TextInput
              style={styles.replyInput}
              placeholder={`Send message to ${currentStory.username}...`}
              placeholderTextColor="rgba(255,255,255,0.6)"
              value={replyText}
              onChangeText={setReplyText}
              onSubmitEditing={handleSendReply}
              onFocus={() => setIsPaused(true)}
              onBlur={() => setIsPaused(false)}
            />
            {replyText.trim().length > 0 && (
              <Pressable onPress={handleSendReply} style={styles.sendBtn}>
                <Ionicons name="arrow-up-circle" size={28} color={colors.primary} />
              </Pressable>
            )}
          </View>

          <Pressable onPress={handleToggleLike} style={styles.iconBtn}>
            <Ionicons
              name={isLiked ? "heart" : "heart-outline"}
              size={28}
              color={isLiked ? colors.heart : "#ffffff"}
            />
          </Pressable>

          <Pressable onPress={() => onReply?.(currentStory, "🔥")} style={styles.iconBtn}>
            <Ionicons name="paper-plane-outline" size={26} color="#ffffff" />
          </Pressable>
        </View>
      </SafeAreaView>
    </Modal>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: "#000000",
  },
  progressContainer: {
    flexDirection: "row",
    gap: 4,
    paddingHorizontal: 12,
    paddingTop: 8,
    paddingBottom: 6,
    zIndex: 10,
  },
  progressBarTrack: {
    flex: 1,
    height: 2.5,
    backgroundColor: "rgba(255,255,255,0.3)",
    borderRadius: 2,
    overflow: "hidden",
  },
  progressBarFill: {
    height: "100%",
    backgroundColor: "#ffffff",
    borderRadius: 2,
  },
  header: {
    flexDirection: "row",
    alignItems: "center",
    justifyContent: "space-between",
    paddingHorizontal: 12,
    paddingVertical: 8,
    zIndex: 10,
  },
  userInfo: {
    flexDirection: "row",
    alignItems: "center",
    gap: 10,
  },
  avatarBorder: {
    width: 34,
    height: 34,
    borderRadius: 17,
    borderWidth: 1.5,
    borderColor: "#ffffff",
    alignItems: "center",
    justifyContent: "center",
  },
  avatar: {
    width: 28,
    height: 28,
    borderRadius: 14,
    backgroundColor: "#262626",
    alignItems: "center",
    justifyContent: "center",
  },
  avatarText: {
    color: "#ffffff",
    fontSize: 11,
    fontWeight: "700",
  },
  username: {
    color: "#ffffff",
    fontSize: 13,
    fontWeight: "700",
  },
  timeAgo: {
    color: "rgba(255,255,255,0.6)",
    fontSize: 12,
  },
  closeBtn: {
    padding: 6,
  },
  mediaContainer: {
    flex: 1,
    borderRadius: 12,
    overflow: "hidden",
    marginHorizontal: 4,
    backgroundColor: "#1a1a1a",
    justifyContent: "center",
    alignItems: "center",
  },
  storyImage: {
    width: "100%",
    height: "100%",
  },
  bottomBar: {
    flexDirection: "row",
    alignItems: "center",
    paddingHorizontal: 12,
    paddingVertical: 10,
    gap: 12,
  },
  replyInputWrap: {
    flex: 1,
    height: 44,
    borderRadius: 22,
    borderWidth: 1,
    borderColor: "rgba(255,255,255,0.35)",
    flexDirection: "row",
    alignItems: "center",
    paddingHorizontal: 16,
    backgroundColor: "rgba(0,0,0,0.4)",
  },
  replyInput: {
    flex: 1,
    color: "#ffffff",
    fontSize: 14,
  },
  sendBtn: {
    marginLeft: 6,
  },
  iconBtn: {
    padding: 4,
  },
});

