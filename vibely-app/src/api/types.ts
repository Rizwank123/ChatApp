export type UUID = string;

export type ApiResponse<T> = {
  data: T;
};

export type PageResponse<T> = {
  data: T[];
  total: number;
  size: number;
  page: number;
};

export type LoginOutput = {
  token: string;
  expires_in: number;
};

export type User = {
  id: UUID;
  user_name: string;
  role?: string;
};

export type Personnel = {
  id: UUID;
  first_name?: string;
  last_name?: string;
  display_name?: string;
  gender?: "MALE" | "FEMALE" | "OTHER";
  email?: string;
  mobile: string;
  role?: "USER" | "ADMIN";
  avatar?: string;
  bio?: string;
  user_id: UUID;
  activation_status?: "ACTIVE" | "DISABLED";
};

export type CheckUsernameAvailabilityOutput = {
  username: string;
  is_available: boolean;
  reason?: string;
};

export type UserRelationship = {
  target_personnel_id: UUID;
  is_following: boolean;
  follows_you: boolean;
  is_mutual: boolean;
  can_chat: boolean;
};

export type FollowUserItem = {
  personnel_id: UUID;
  user_id: UUID;
  user_name: string;
  display_name: string;
  avatar?: string;
  bio?: string;
  is_following: boolean;
  follows_you: boolean;
  is_mutual: boolean;
  created_at: string;
};

export type UserProfileSearchResult = {
  user_id: UUID;
  personnel_id: UUID;
  user_name: string;
  display_name: string;
  avatar?: string;
  bio?: string;
  is_following: boolean;
  follows_you: boolean;
  is_mutual: boolean;
};

export type FilterInput = {
  fields: Array<{
    field: string;
    operator:
      | "eq"
      | "neq"
      | "gt"
      | "gte"
      | "lt"
      | "lte"
      | "in"
      | "nin"
      | "like"
      | "nlike"
      | "ilike"
      | "nilike"
      | "isnull"
      | "notnull"
      | "between";
    value: string | number | boolean | string[];
  }>;
  sort_keys?: Array<{
    field: string;
    direction: "asc" | "desc";
  }>;
};

export type RegisterInput = {
  first_name: string;
  last_name: string;
  user_name: string;
  role: "USER" | "ADMIN";
  password: string;
};

export type LoginInput = {
  username: string;
  password: string;
};

export type MediaItem = {
  id: UUID;
  media_type: "IMAGE" | "VIDEO";
  url: string;
  thumbnail_url?: string;
  width?: number;
  height?: number;
  mime_type: string;
  position: number;
};

export type InitiateUploadResponse = {
  upload_id: UUID;
  presigned_url: string;
  object_key: string;
  bucket: string;
  media_type: "IMAGE" | "VIDEO";
  cdn_url: string;
};

export type UserSearchResult = {
  user_id: UUID;
  personnel_id: UUID;
  user_name: string;
  display_name: string;
  avatar?: string;
  bio?: string;
  follower_count: number;
  is_following: boolean;
  follows_you: boolean;
  match_score?: number;
};

export type HashtagSearchResult = {
  id: UUID;
  name: string;
  post_count: number;
  match_score?: number;
};

export type UnifiedSearchResult = {
  users: UserSearchResult[];
  posts: Post[];
  hashtags: HashtagSearchResult[];
};

export type Post = {
  id: UUID;
  personnel_id: UUID;
  creator_name?: string;
  creator_avatar?: string;
  caption: string;
  media_url: string[];
  media?: MediaItem[];
  location?: string;
  visibility: "PUBLIC" | "PRIVATE";
  likes_count?: number;
  comments_count?: number;
  views_count?: number;
  shares_count?: number;
  likes?: number;
  comments?: number;
  is_liked?: boolean;
  is_saved?: boolean;
  created_at?: string;
  hashtags?: string[];
};

export type Notification = {
  id: UUID;
  recipient_id: UUID;
  actor_id: UUID;
  actor_name?: string;
  actor_avatar?: string;
  type: "LIKE" | "COMMENT" | "FOLLOW" | "MENTION";
  post_id?: UUID;
  comment_id?: UUID;
  is_read: boolean;
  created_at: string;
};

export type UserCounters = {
  personnel_id: UUID;
  follower_count: number;
  following_count: number;
  post_count: number;
};

export type FeedPage = {
  items: Post[];
  next_cursor?: string;
  has_more: boolean;
};

export type CreatePostInput = {
  caption: string;
  media_url?: string[];
  media_ids?: UUID[];
  location?: string;
  visibility?: "PUBLIC" | "PRIVATE";
  hashtags?: string[];
};

export type Comment = {
  id: UUID;
  personnel_id: UUID;
  post_id: UUID;
  parent_id?: UUID;
  content: string;
  created_at?: string;
};

export type Friendship = {
  id: UUID;
  personnel_id: UUID;
  friend_id: UUID;
  status: "PENDING" | "ACCEPTED" | "BLOCKED";
};

export type Conversation = {
  id: UUID;
  name?: string;
  conversation_type: "DIRECT" | "GROUP";
  created_at?: string;
};

export type Message = {
  id: UUID;
  conversation_id: UUID;
  sender_id: UUID;
  message_type: "TEXT" | "IMAGE" | "FILE";
  content: string;
  created_at?: string;
};

export type AppUser = {
  token: string;
  expiresIn: number;
  userName: string;
  personnelId?: UUID;
};
