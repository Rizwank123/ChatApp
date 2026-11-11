import { Platform } from "react-native";

import type {
  ApiResponse,
  CheckUsernameAvailabilityOutput,
  Comment,
  Conversation,
  CreatePostInput,
  FollowUserItem,
  Friendship,
  HashtagSearchResult,
  InitiateUploadResponse,
  LoginInput,
  LoginOutput,
  Message,
  PageResponse,
  Personnel,
  Post,
  RegisterInput,
  UnifiedSearchResult,
  User,
  UserCounters,
  UserProfileSearchResult,
  UserRelationship,
  UserSearchResult,
  UUID,
} from "./types";

const defaultHost = Platform.OS === "android" ? "192.168.1.36" : "localhost";
const baseUrl =
  process.env.EXPO_PUBLIC_API_URL ?? `http://${defaultHost}:7700/api/v1`;
const socketUrl =
  process.env.EXPO_PUBLIC_WS_URL ?? `ws://${defaultHost}:7700/ws`;

type RequestOptions = {
  token?: string;
  method?: "GET" | "POST" | "PUT" | "DELETE" | "PATCH";
  body?: unknown;
};

class ApiError extends Error {
  status: number;

  constructor(status: number, message: string) {
    super(message);
    this.status = status;
  }
}

async function request<T>(path: string, options: RequestOptions = {}): Promise<T> {
  const response = await fetch(`${baseUrl}${path}`, {
    method: options.method ?? "GET",
    headers: {
      Accept: "application/json",
      "Content-Type": "application/json",
      ...(options.token ? { Authorization: `Bearer ${options.token}` } : {}),
    },
    body: options.body ? JSON.stringify(options.body) : undefined,
  });

  if (response.status === 204) {
    return undefined as T;
  }

  const text = await response.text();
  const json = text ? JSON.parse(text) : {};

  if (!response.ok) {
    throw new ApiError(response.status, json.message ?? "Request failed");
  }

  return json as T;
}

const qs = (page = 0, limit = 20) => `page=${page}&limit=${limit}&size=${limit}`;

export const apiConfig = {
  baseUrl,
  socketUrl,
};

export const api = {
  async register(input: RegisterInput) {
    return request<ApiResponse<User>>("/users", { method: "POST", body: input });
  },
  async checkUsername(username: string) {
    return request<ApiResponse<CheckUsernameAvailabilityOutput>>(
      `/users/check-username?username=${encodeURIComponent(username)}`,
    );
  },
  async login(input: LoginInput) {
    return request<ApiResponse<LoginOutput>>("/users/login", {
      method: "POST",
      body: input,
    });
  },
  async getFeed(token: string, page = 0, size = 20) {
    return request<PageResponse<Post>>(`/posts/feed?${qs(page, size)}`, { token });
  },
  async getHybridFeed(token: string, cursor = "", limit = 20) {
    const query = cursor ? `cursor=${encodeURIComponent(cursor)}&limit=${limit}` : `limit=${limit}`;
    return request<ApiResponse<{ items: Post[]; next_cursor?: string; has_more: boolean }>>(`/posts/feed?${query}`, { token });
  },
  async savePost(token: string, postId: UUID) {
    return request<ApiResponse<{ message: string }>>(`/posts/${postId}/save`, {
      token,
      method: "POST",
    });
  },
  async unsavePost(token: string, postId: UUID) {
    return request<ApiResponse<{ message: string }>>(`/posts/${postId}/save`, {
      token,
      method: "DELETE",
    });
  },
  async followSocial(token: string, targetId: UUID) {
    return request<ApiResponse<{ message: string; relationship_status: string }>>(`/social/follow/${targetId}`, {
      token,
      method: "POST",
    });
  },
  async unfollowSocial(token: string, targetId: UUID) {
    return request<ApiResponse<{ message: string }>>(`/social/follow/${targetId}`, {
      token,
      method: "DELETE",
    });
  },
  async getRelationship(token: string, targetId: UUID) {
    return request<ApiResponse<UserRelationship>>(`/social/relationship/${targetId}`, {
      token,
    });
  },
  async canChat(token: string, targetId: UUID) {
    return request<
      ApiResponse<{
        can_chat: boolean;
        is_mutual: boolean;
        is_following: boolean;
        follows_you: boolean;
        reason?: string;
      }>
    >(`/social/can-chat/${targetId}`, { token });
  },
  async getUserCounters(token: string, personnelId?: UUID) {
    const path = personnelId ? `/social/counters/${personnelId}` : "/social/counters";
    return request<ApiResponse<UserCounters>>(path, { token });
  },
  async getFollowers(token: string, targetId: UUID, page = 0, size = 20, query = "") {
    const q = query.trim() ? `&q=${encodeURIComponent(query.trim())}` : "";
    return request<PageResponse<FollowUserItem>>(
      `/social/followers/${targetId}?offset=${page * size}&limit=${size}${q}`,
      { token },
    );
  },
  async getFollowing(token: string, targetId: UUID, page = 0, size = 20, query = "") {
    const q = query.trim() ? `&q=${encodeURIComponent(query.trim())}` : "";
    return request<PageResponse<FollowUserItem>>(
      `/social/following/${targetId}?offset=${page * size}&limit=${size}${q}`,
      { token },
    );
  },
  async searchUsers(token: string, query: string, limit = 20) {
    return request<ApiResponse<UserProfileSearchResult[]>>(
      `/social/search?q=${encodeURIComponent(query)}&limit=${limit}`,
      { token },
    );
  },
  async blockUser(token: string, targetId: UUID) {
    return request<ApiResponse<{ message: string }>>(`/social/block/${targetId}`, {
      token,
      method: "POST",
    });
  },
  async unblockUser(token: string, targetId: UUID) {
    return request<ApiResponse<{ message: string }>>(`/social/block/${targetId}`, {
      token,
      method: "DELETE",
    });
  },
  async getNotifications(token: string) {
    return request<PageResponse<any>>("/notifications", { token });
  },
  async markNotificationRead(token: string, notifId: UUID) {
    return request<ApiResponse<{ message: string }>>(`/notifications/${notifId}/read`, {
      token,
      method: "PATCH",
    });
  },
  async getPersonnel(token: string, personnelId: UUID) {
    return request<ApiResponse<Personnel>>(`/personnel/${personnelId}`, { token });
  },
  async getUser(token: string, userId: UUID) {
    return request<ApiResponse<User>>(`/users/${userId}`, { token });
  },
  async findUserByID(token: string, userId: UUID) {
    return request<ApiResponse<User>>(`/users/${userId}`, { token });
  },
  async findUserByUsername(token: string, username: string) {
    return request<ApiResponse<User>>(`/users/${encodeURIComponent(username)}`, { token });
  },
  async findPersonnelByUserID(token: string, userId: UUID) {
    return request<ApiResponse<Personnel>>(`/personnel/user/${userId}`, { token }).catch(() =>
      request<ApiResponse<Personnel>>(`/personnel/user/id?id=${userId}`, { token }),
    );
  },
  async getUserPosts(token: string, userId: UUID, page = 0, size = 20) {
    return request<PageResponse<Post>>(`/users/${userId}/posts?${qs(page, size)}`, {
      token,
    });
  },
  async createPost(token: string, input: CreatePostInput) {
    return request<ApiResponse<null>>("/posts", {
      token,
      method: "POST",
      body: input,
    });
  },
  async updatePost(token: string, postId: UUID, input: Partial<CreatePostInput>) {
    return request<ApiResponse<null>>(`/posts/${postId}`, {
      token,
      method: "PUT",
      body: input,
    });
  },
  async deletePost(token: string, postId: UUID) {
    return request<void>(`/posts/${postId}`, { token, method: "DELETE" });
  },
  async likePost(token: string, postId: UUID) {
    return request<void>(`/posts/${postId}/like`, { token, method: "POST" });
  },
  async unlikePost(token: string, postId: UUID) {
    return request<void>(`/posts/${postId}/unlike`, { token, method: "POST" });
  },
  async getComments(token: string, postId: UUID, page = 0, size = 50) {
    return request<PageResponse<Comment>>(
      `/posts/${postId}/comments?${qs(page, size)}`,
      { token },
    );
  },
  async createComment(token: string, postId: UUID, content: string, parentId?: UUID) {
    return request<ApiResponse<null>>(`/posts/${postId}/comments`, {
      token,
      method: "POST",
      body: { content, parent_id: parentId },
    });
  },
  async deleteComment(token: string, postId: UUID, commentId: UUID) {
    return request<void>(`/posts/${postId}/comments/${commentId}`, {
      token,
      method: "DELETE",
    });
  },
  async follow(token: string, userId: UUID) {
    return request<void>(`/users/${userId}/follow`, { token, method: "POST" });
  },
  async unfollow(token: string, userId: UUID) {
    return request<void>(`/users/${userId}/unfollow`, { token, method: "POST" });
  },
  async filterPersonnel(token: string, query: string, page = 0, size = 20) {
    const body = query.trim()
      ? {
        fields: [
          {
            field: "first_name",
            operator: "ilike",
            value: query.trim(),
          },
        ],
        sort_keys: [{ field: "first_name", direction: "asc" }],
      }
      : {
        fields: [],
        sort_keys: [{ field: "first_name", direction: "asc" }],
      };

    return request<PageResponse<Personnel>>(`/personnel/filter?${qs(page, size)}`, {
      token,
      method: "POST",
      body,
    });
  },
  async getFriendships(token: string) {
    return request<ApiResponse<Friendship[]>>("/friends", { token });
  },
  async sendFriendRequest(token: string, friendUsername: string) {
    return request<ApiResponse<Friendship>>("/friends", {
      token,
      method: "POST",
      body: { friend_username: friendUsername },
    });
  },
  async acceptFriendRequest(token: string, requestId: UUID) {
    return request<ApiResponse<null>>(`/friends/${requestId}/accept`, {
      token,
      method: "POST",
    });
  },
  async getConversations(token: string) {
    return request<ApiResponse<Conversation[]>>("/conversations", { token });
  },
  async createConversation(
    token: string,
    participantIds: UUID[],
    conversationType: "DIRECT" | "GROUP",
    name = "",
  ) {
    return request<ApiResponse<Conversation>>("/conversations", {
      token,
      method: "POST",
      body: {
        name,
        conversation_type: conversationType,
        participant_ids: participantIds,
      },
    });
  },
  async getMessages(token: string, conversationId: UUID, page = 0, size = 50) {
    return request<PageResponse<Message>>(
      `/messages/${conversationId}?${qs(page, size)}`,
      { token },
    );
  },
  async sendMessage(token: string, conversationId: UUID, content: string) {
    return request<ApiResponse<Message>>("/messages", {
      token,
      method: "POST",
      body: {
        conversation_id: conversationId,
        message_type: "TEXT",
        content,
      },
    });
  },
  async markMessage(token: string, messageId: UUID, status: "DELIVERED" | "READ") {
    return request<ApiResponse<null>>(`/messages/${messageId}`, {
      token,
      method: "PUT",
      body: { status },
    });
  },
  async getExplorePosts(token: string, page = 0, size = 20) {
    return request<PageResponse<Post>>(`/posts/explore?${qs(page, size)}`, { token });
  },
  async recordPostView(token: string, postId: UUID) {
    return request<ApiResponse<{ message: string }>>(`/posts/${postId}/view`, {
      token,
      method: "POST",
    });
  },
  async recordPostShare(token: string, postId: UUID) {
    return request<ApiResponse<{ message: string }>>(`/posts/${postId}/share`, {
      token,
      method: "POST",
    });
  },
  async initiateUpload(
    token: string,
    fileName: string,
    mimeType: string,
    sizeBytes: number,
    purpose: "POST_IMAGE" | "POST_VIDEO" | "AVATAR" | "MESSAGE_ATTACHMENT" = "POST_IMAGE",
  ) {
    return request<ApiResponse<InitiateUploadResponse>>("/media/uploads", {
      token,
      method: "POST",
      body: {
        file_name: fileName,
        mime_type: mimeType,
        size_bytes: sizeBytes,
        purpose,
      },
    });
  },
  async completeUpload(token: string, uploadId: UUID) {
    return request<ApiResponse<any>>(`/media/uploads/${uploadId}/complete`, {
      token,
      method: "POST",
    });
  },
  async uploadFileToUrl(presignedUrl: string, blob: Blob, mimeType: string) {
    const res = await fetch(presignedUrl, {
      method: "PUT",
      headers: {
        "Content-Type": mimeType,
      },
      body: blob,
    });
    if (!res.ok) {
      throw new Error(`Upload failed with status ${res.status}`);
    }
  },
  async unifiedSearch(token: string, query: string, limit = 10) {
    return request<ApiResponse<UnifiedSearchResult>>(
      `/search?q=${encodeURIComponent(query)}&limit=${limit}`,
      { token },
    );
  },
  async searchUsersAdvanced(token: string, query: string, page = 0, size = 20) {
    return request<PageResponse<UserSearchResult>>(
      `/search/users?q=${encodeURIComponent(query)}&${qs(page, size)}`,
      { token },
    );
  },
  async searchPosts(token: string, query: string, page = 0, size = 20) {
    return request<PageResponse<Post>>(
      `/search/posts?q=${encodeURIComponent(query)}&${qs(page, size)}`,
      { token },
    );
  },
  async searchHashtags(token: string, query: string, page = 0, size = 20) {
    return request<PageResponse<HashtagSearchResult>>(
      `/search/hashtags?q=${encodeURIComponent(query)}&${qs(page, size)}`,
      { token },
    );
  },
  async getHashtagPosts(token: string, tag: string, page = 0, size = 20) {
    return request<PageResponse<Post>>(
      `/hashtags/${encodeURIComponent(tag)}/posts?${qs(page, size)}`,
      { token },
    );
  },
  async getHashtagDetails(token: string, tag: string) {
    return request<ApiResponse<HashtagSearchResult>>(
      `/hashtags/${encodeURIComponent(tag)}`,
      { token },
    );
  },
};
