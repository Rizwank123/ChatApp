import React, { useCallback, useEffect, useState } from "react";
import { FlatList, Image, Pressable, StyleSheet, Text, View } from "react-native";

import { api } from "../api/client";
import type { Friendship, Personnel } from "../api/types";
import { Button } from "../components/Button";
import { Field } from "../components/Field";
import { useApp } from "../state/AppContext";
import { colors } from "../theme/colors";

export function SocialScreen() {
  const { session, setSession, showToast } = useApp();
  const [query, setQuery] = useState("");
  const [people, setPeople] = useState<Personnel[]>([]);
  const [friendships, setFriendships] = useState<Friendship[]>([]);
  const [skippedIds, setSkippedIds] = useState<Set<string>>(new Set());
  const [loading, setLoading] = useState(false);

  const visiblePeople = people.filter(
    (person) => person.id !== session?.personnelId && !skippedIds.has(person.id),
  );

  const loadFriendships = useCallback(async () => {
    if (!session) {
      return;
    }
    const result = await api.getFriendships(session.token);
    setFriendships(result.data ?? []);
  }, [session]);

  const searchPeople = useCallback(async () => {
    if (!session) {
      return;
    }
    setLoading(true);
    try {
      const result = await api.filterPersonnel(session.token, query);
      setPeople(result.data ?? []);
    } catch (error) {
      showToast({
        message: error instanceof Error ? error.message : "Could not search people",
        kind: "error",
      });
    } finally {
      setLoading(false);
    }
  }, [query, session, showToast]);

  useEffect(() => {
    loadFriendships().catch(() => undefined);
  }, [loadFriendships]);

  useEffect(() => {
    if (!session) {
      return;
    }
    api
      .filterPersonnel(session.token, "")
      .then((result) => setPeople(result.data ?? []))
      .catch(() => undefined);
  }, [session]);

  async function addPerson(person: Personnel) {
    if (!session) {
      return;
    }
    try {
      await api.sendFriendRequest(session.token, person.mobile);
      setSkippedIds((current) => new Set(current).add(person.id));
      showToast({
        message: `Friend request sent to ${person.first_name || person.mobile}`,
        kind: "success",
      });
      loadFriendships();
    } catch (error) {
      showToast({
        message: error instanceof Error ? error.message : "Could not add user",
        kind: "error",
      });
    }
  }

  function skipPerson(person: Personnel) {
    setSkippedIds((current) => new Set(current).add(person.id));
  }

  async function accept(item: Friendship) {
    if (!session) {
      return;
    }
    await api.acceptFriendRequest(session.token, item.id);
    loadFriendships();
  }

  return (
    <FlatList
      contentContainerStyle={styles.content}
      data={visiblePeople}
      keyExtractor={(item) => item.id}
      keyboardShouldPersistTaps="handled"
      ListHeaderComponent={
        <View style={styles.header}>
          <View>
            <Text style={styles.title}>People</Text>
            <Text style={styles.subtitle}>{session?.userName}</Text>
          </View>

          <View style={styles.panel}>
            <Field
              label="Search users"
              value={query}
              onChangeText={setQuery}
              placeholder="Search by first name"
              returnKeyType="search"
              onSubmitEditing={searchPeople}
            />
            <Button
              label={loading ? "Searching..." : "Search"}
              onPress={searchPeople}
              disabled={loading}
            />
          </View>
        </View>
      }
      ListEmptyComponent={
        <Text style={styles.empty}>
          {loading ? "Searching..." : "No matching users to show."}
        </Text>
      }
      renderItem={({ item }) => (
        <View style={styles.personCard}>
          <View style={styles.personTop}>
            {item.avatar ? (
              <Image source={{ uri: item.avatar }} style={styles.avatarImage} />
            ) : (
              <View style={styles.avatarFallback}>
                <Text style={styles.avatarText}>
                  {(item.first_name || item.mobile || "?").slice(0, 2)}
                </Text>
              </View>
            )}
            <View style={styles.personInfo}>
              <Text style={styles.personName}>
                {[item.first_name, item.last_name].filter(Boolean).join(" ") ||
                  item.mobile}
              </Text>
              <Text style={styles.personMeta}>{item.mobile}</Text>
              {item.email ? <Text style={styles.personMeta}>{item.email}</Text> : null}
            </View>
          </View>
          <View style={styles.actions}>
            <Button label="Add" onPress={() => addPerson(item)} style={styles.flex} />
            <Button
              label="Skip"
              variant="secondary"
              onPress={() => skipPerson(item)}
              style={styles.flex}
            />
          </View>
        </View>
      )}
      ListFooterComponent={
        <View style={styles.footer}>
          <Text style={styles.sectionTitle}>Friend requests</Text>
          {friendships.length === 0 ? (
            <Text style={styles.emptySmall}>No friend requests yet.</Text>
          ) : (
            friendships.map((item) => (
              <View key={item.id} style={styles.friendship}>
                <View style={styles.friendshipText}>
                  <Text style={styles.friendTitle}>
                    {item.personnel_id.slice(0, 8)} to {item.friend_id.slice(0, 8)}
                  </Text>
                  <Text style={styles.status}>{item.status}</Text>
                </View>
                {item.status === "PENDING" ? (
                  <Button
                    label="Accept"
                    variant="secondary"
                    onPress={() => accept(item)}
                  />
                ) : null}
              </View>
            ))
          )}
          <Pressable onPress={() => setSession(undefined)} style={styles.signOut}>
            <Text style={styles.signOutText}>Sign out</Text>
          </Pressable>
        </View>
      }
    />
  );
}

const styles = StyleSheet.create({
  content: {
    padding: 16,
    gap: 14,
  },
  header: {
    gap: 14,
  },
  title: {
    color: colors.text,
    fontSize: 30,
    fontWeight: "900",
  },
  subtitle: {
    color: colors.textDim,
  },
  panel: {
    gap: 10,
    backgroundColor: colors.cardSubtle,
    borderWidth: 1,
    borderColor: colors.border,
    borderRadius: 8,
    padding: 14,
  },
  empty: {
    color: colors.textDim,
    textAlign: "center",
    padding: 24,
  },
  emptySmall: {
    color: colors.textDim,
    paddingVertical: 8,
  },
  personCard: {
    backgroundColor: colors.cardSubtle,
    borderRadius: 8,
    borderWidth: 1,
    borderColor: colors.border,
    padding: 14,
    gap: 14,
  },
  personTop: {
    flexDirection: "row",
    alignItems: "center",
    gap: 12,
  },
  avatarImage: {
    width: 54,
    height: 54,
    borderRadius: 27,
    backgroundColor: colors.soft,
  },
  avatarFallback: {
    width: 54,
    height: 54,
    borderRadius: 27,
    alignItems: "center",
    justifyContent: "center",
    backgroundColor: colors.cardSubtle,
  },
  avatarText: {
    color: colors.text,
    fontWeight: "900",
    textTransform: "uppercase",
  },
  personInfo: {
    flex: 1,
    gap: 3,
  },
  personName: {
    color: colors.text,
    fontSize: 17,
    fontWeight: "900",
  },
  personMeta: {
    color: colors.textDim,
  },
  actions: {
    flexDirection: "row",
    gap: 10,
  },
  flex: {
    flex: 1,
  },
  footer: {
    gap: 10,
    paddingTop: 4,
  },
  sectionTitle: {
    color: colors.text,
    fontSize: 18,
    fontWeight: "900",
  },
  friendship: {
    flexDirection: "row",
    alignItems: "center",
    backgroundColor: colors.cardSubtle,
    borderRadius: 8,
    borderWidth: 1,
    borderColor: colors.border,
    padding: 12,
    gap: 10,
  },
  friendshipText: {
    flex: 1,
  },
  friendTitle: {
    color: colors.text,
    fontWeight: "800",
  },
  status: {
    color: colors.textDim,
  },
  signOut: {
    minHeight: 46,
    borderRadius: 8,
    alignItems: "center",
    justifyContent: "center",
    backgroundColor: colors.cardElevated,
    borderWidth: 1,
    borderColor: colors.border,
  },
  signOutText: {
    color: colors.danger,
    fontWeight: "800",
  },
});
