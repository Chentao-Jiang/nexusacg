class UserModel {
  final String id;
  final String? phone;
  final String? email;
  final String nickname;
  final String? avatarUrl;
  final String bio;
  final String role;

  UserModel({
    required this.id,
    this.phone,
    this.email,
    required this.nickname,
    this.avatarUrl,
    this.bio = '',
    this.role = 'user',
  });

  factory UserModel.fromJson(Map<String, dynamic> json) {
    return UserModel(
      id: json['id'] as String,
      phone: json['phone'] as String?,
      email: json['email'] as String?,
      nickname: json['nickname'] as String? ?? '',
      avatarUrl: json['avatar_url'] as String?,
      bio: json['bio'] as String? ?? '',
      role: json['role'] as String? ?? 'user',
    );
  }
}

class ProductModel {
  final String id;
  final String name;
  final String description;
  final double price;
  final double? originalPrice;
  final String zone;
  final String sourceType;
  final List<String> images;
  final int stock;
  final String status;
  final List<String> tags;
  final String? characterName;
  final String? animeName;

  ProductModel({
    required this.id,
    required this.name,
    this.description = '',
    required this.price,
    this.originalPrice,
    required this.zone,
    required this.sourceType,
    this.images = const [],
    this.stock = 0,
    this.status = 'active',
    this.tags = const [],
    this.characterName,
    this.animeName,
  });

  factory ProductModel.fromJson(Map<String, dynamic> json) {
    return ProductModel(
      id: json['id'] as String,
      name: json['name'] as String,
      description: json['description'] as String? ?? '',
      price: (json['price'] as num).toDouble(),
      originalPrice: json['original_price'] != null ? (json['original_price'] as num).toDouble() : null,
      zone: json['zone'] as String,
      sourceType: json['source_type'] as String,
      images: json['images'] != null ? List<String>.from(json['images']) : [],
      stock: json['stock'] as int? ?? 0,
      status: json['status'] as String? ?? 'active',
      tags: json['tags'] != null ? List<String>.from(json['tags']) : [],
      characterName: json['character_name'] as String?,
      animeName: json['anime_name'] as String?,
    );
  }
}

class PostModel {
  final String id;
  final String userId;
  final String title;
  final String content;
  final List<String> images;
  final String? videoUrl;
  final String type;
  final List<String> tags;
  final int likeCount;
  final int commentCount;
  final String status;
  final DateTime createdAt;
  final UserModel? author;

  PostModel({
    required this.id,
    required this.userId,
    this.title = '',
    required this.content,
    this.images = const [],
    this.videoUrl,
    this.type = 'text',
    this.tags = const [],
    this.likeCount = 0,
    this.commentCount = 0,
    this.status = 'approved',
    required this.createdAt,
    this.author,
  });

  factory PostModel.fromJson(Map<String, dynamic> json) {
    return PostModel(
      id: json['id'] as String,
      userId: json['user_id'] as String,
      title: json['title'] as String? ?? '',
      content: json['content'] as String,
      images: json['images'] != null ? List<String>.from(json['images']) : [],
      videoUrl: json['video_url'] as String?,
      type: json['type'] as String? ?? 'text',
      tags: json['tags'] != null ? List<String>.from(json['tags']) : [],
      likeCount: json['like_count'] as int? ?? 0,
      commentCount: json['comment_count'] as int? ?? 0,
      status: json['status'] as String? ?? 'approved',
      createdAt: DateTime.parse(json['created_at'] as String),
      author: json['author'] != null ? UserModel.fromJson(json['author']) : null,
    );
  }
}

class EventModel {
  final String id;
  final String name;
  final String description;
  final String? coverUrl;
  final DateTime startTime;
  final DateTime endTime;
  final String address;
  final double? latitude;
  final double? longitude;
  final String status;

  EventModel({
    required this.id,
    required this.name,
    this.description = '',
    this.coverUrl,
    required this.startTime,
    required this.endTime,
    required this.address,
    this.latitude,
    this.longitude,
    this.status = 'upcoming',
  });

  factory EventModel.fromJson(Map<String, dynamic> json) {
    return EventModel(
      id: json['id'] as String,
      name: json['name'] as String,
      description: json['description'] as String? ?? '',
      coverUrl: json['cover_url'] as String?,
      startTime: DateTime.parse(json['start_time'] as String),
      endTime: DateTime.parse(json['end_time'] as String),
      address: json['address'] as String,
      latitude: json['latitude'] != null ? (json['latitude'] as num).toDouble() : null,
      longitude: json['longitude'] != null ? (json['longitude'] as num).toDouble() : null,
      status: json['status'] as String? ?? 'upcoming',
    );
  }
}
