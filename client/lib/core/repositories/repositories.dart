import 'package:nexusacg/core/models/models.dart';
import 'package:nexusacg/core/network/api_client.dart';

class ProductRepository {
  final ApiClient _api = ApiClient();

  Future<List<ProductModel>> getProducts({
    String? zone,
    String? keyword,
    int page = 1,
    int pageSize = 20,
    String sort = 'newest',
  }) async {
    final params = <String, dynamic>{'page': page, 'page_size': pageSize, 'sort': sort};
    if (zone != null) params['zone'] = zone;
    if (keyword != null && keyword.isNotEmpty) params['keyword'] = keyword;

    final response = await _api.get('/products', queryParameters: params);
    final data = response.data;
    if (data['code'] == 0 && data['data'] != null) {
      final items = data['data']['items'] as List;
      return items.map((e) => ProductModel.fromJson(e as Map<String, dynamic>)).toList();
    }
    return [];
  }

  Future<ProductModel?> getProduct(String id) async {
    final response = await _api.get('/products/$id');
    final data = response.data;
    if (data['code'] == 0 && data['data'] != null) {
      return ProductModel.fromJson(data['data'] as Map<String, dynamic>);
    }
    return null;
  }
}

class PostRepository {
  final ApiClient _api = ApiClient();

  Future<List<PostModel>> getPosts({
    int page = 1,
    int pageSize = 20,
    String? tag,
  }) async {
    final params = <String, dynamic>{'page': page, 'page_size': pageSize};
    if (tag != null) params['tag'] = tag;

    final response = await _api.get('/posts', queryParameters: params);
    final data = response.data;
    if (data['code'] == 0 && data['data'] != null) {
      final items = data['data']['items'] as List;
      return items.map((e) => PostModel.fromJson(e as Map<String, dynamic>)).toList();
    }
    return [];
  }

  Future<void> likePost(String postId) async {
    await _api.post('/posts/$postId/like');
  }

  Future<void> unlikePost(String postId) async {
    await _api.delete('/posts/$postId/like');
  }

  Future<Map<String, dynamic>?> createPost({
    required String title,
    required String content,
    List<String> images = const [],
    String? videoUrl,
    List<String> tags = const [],
    String? visibility,
    String? groupId,
  }) async {
    final body = <String, dynamic>{'content': content};
    if (title.isNotEmpty) body['title'] = title;
    if (images.isNotEmpty) body['images'] = images;
    if (videoUrl != null) body['video_url'] = videoUrl;
    if (groupId != null) body['group_id'] = groupId;
    if (tags.isNotEmpty) body['tags'] = tags;
    if (visibility != null) body['visibility'] = visibility;
    final response = await _api.post('/posts', data: body);
    return response.data;
  }

  Future<Map<String, dynamic>?> updatePost(String postId, {
    String? title,
    String? content,
    List<String>? images,
    String? videoUrl,
    List<String>? tags,
    String? visibility,
    String? groupId,
  }) async {
    final body = <String, dynamic>{};
    if (title != null) body['title'] = title;
    if (content != null) body['content'] = content;
    if (images != null) body['images'] = images;
    if (videoUrl != null) body['video_url'] = videoUrl;
    if (groupId != null) body['group_id'] = groupId;
    if (tags != null) body['tags'] = tags;
    if (visibility != null) body['visibility'] = visibility;
    final response = await _api.put('/posts/$postId', data: body);
    return response.data;
  }
}

class EventRepository {
  final ApiClient _api = ApiClient();

  Future<({List<EventModel> items, int total})> getEvents({
    int page = 1,
    int pageSize = 20,
    String? status,
  }) async {
    final params = <String, dynamic>{'page': page, 'page_size': pageSize};
    if (status != null) params['status'] = status;

    final response = await _api.get('/events', queryParameters: params);
    final data = response.data;
    if (data['code'] == 0 && data['data'] != null) {
      final items = data['data']['items'] as List;
      final total = data['data']['total'] as int? ?? 0;
      return (
        items: items.map((e) => EventModel.fromJson(e as Map<String, dynamic>)).toList(),
        total: total,
      );
    }
    return (items: <EventModel>[], total: 0);
  }

  Future<EventModel?> getEvent(String id) async {
    final response = await _api.get('/events/$id');
    final data = response.data;
    if (data['code'] == 0 && data['data'] != null) {
      return EventModel.fromJson(data['data'] as Map<String, dynamic>);
    }
    return null;
  }

  Future<EventModel?> createEvent({
    required String name,
    required String description,
    required DateTime startTime,
    required DateTime endTime,
    required String address,
    double? latitude,
    double? longitude,
    String? coverUrl,
  }) async {
    final body = <String, dynamic>{
      'name': name,
      'description': description,
      'start_time': startTime.toIso8601String(),
      'end_time': endTime.toIso8601String(),
      'address': address,
    };
    if (latitude != null) body['latitude'] = latitude;
    if (longitude != null) body['longitude'] = longitude;
    if (coverUrl != null) body['cover_url'] = coverUrl;

    final response = await _api.post('/events', data: body);
    final data = response.data;
    if (data['code'] == 0 && data['data'] != null) {
      return EventModel.fromJson(data['data'] as Map<String, dynamic>);
    }
    return null;
  }
}

class AuthRepository {
  final ApiClient _api = ApiClient();

  Future<Map<String, dynamic>> register({
    required String nickname,
    required String password,
    String? phone,
    String? email,
  }) async {
    final body = <String, dynamic>{'nickname': nickname, 'password': password};
    if (phone != null) body['phone'] = phone;
    if (email != null) body['email'] = email;

    final response = await _api.post('/auth/register', data: body);
    return response.data;
  }

  Future<void> sendSmsCode(String phone) async {
    final response = await _api.post('/auth/sms/send', data: {'phone': phone});
    final data = response.data;
    if (data is! Map || data['code'] != 0) {
      throw Exception(data is Map ? (data['message'] ?? '发送失败') : '发送失败');
    }
  }

  Future<Map<String, dynamic>> smsLogin(String phone, String code, String password, String nickname) async {
    final response = await _api.post('/auth/sms/login', data: {
      'phone': phone,
      'code': code,
      'password': password,
      'nickname': nickname,
    });
    final data = response.data;
    if (data is! Map || data['code'] != 0) {
      throw Exception(data is Map ? (data['message'] ?? '注册失败') : '注册失败');
    }
    return Map<String, dynamic>.from(data);
  }

  Future<Map<String, dynamic>> login({
    String? phone,
    String? email,
    required String password,
  }) async {
    final body = <String, dynamic>{'password': password};
    if (phone != null) body['phone'] = phone;
    if (email != null) body['email'] = email;

    final response = await _api.post('/auth/login', data: body);
    return response.data;
  }

  Future<Map<String, dynamic>> getCurrentUser() async {
    final response = await _api.post('/auth/me');
    return response.data;
  }

  Future<Map<String, dynamic>> updateProfile({String? nickname, String? bio, String? avatarUrl}) async {
    final body = <String, dynamic>{};
    if (nickname != null) body['nickname'] = nickname;
    if (bio != null) body['bio'] = bio;
    if (avatarUrl != null) body['avatar_url'] = avatarUrl;
    final response = await _api.post('/auth/profile', data: body);
    final data = response.data;
    if (data is! Map) {
      throw Exception('服务器返回异常: $data');
    }
    return Map<String, dynamic>.from(data);
  }

  Future<String?> getQQAuthUrl() async {
    final response = await _api.get('/auth/qq/authorize');
    final data = response.data;
    if (data['code'] == 0 && data['data'] != null) {
      return data['data']['redirect_url'] as String?;
    }
    return null;
  }

  Future<void> resendEmailVerification(String email) async {
    final response = await _api.post('/auth/email/resend', data: {'email': email});
    final data = response.data;
    if (data is! Map || data['code'] != 0) {
      throw Exception(data is Map ? (data['message'] ?? '发送失败') : '发送失败');
    }
  }

  Future<Map<String, dynamic>?> qqCallback(String code) async {
    final response = await _api.get('/auth/qq/callback', queryParameters: {'code': code});
    return response.data;
  }
}

class OrderRepository {
  final ApiClient _api = ApiClient();

  Future<({List<OrderModel> items, int total})> getOrders({
    int page = 1,
    int pageSize = 20,
    String? status,
  }) async {
    final params = <String, dynamic>{'page': page, 'page_size': pageSize};
    if (status != null) params['status'] = status;

    final response = await _api.get('/orders', queryParameters: params);
    final data = response.data;
    if (data['code'] == 0 && data['data'] != null) {
      final items = data['data']['items'] as List;
      final total = data['data']['total'] as int? ?? 0;
      return (
        items: items.map((e) => OrderModel.fromJson(e as Map<String, dynamic>)).toList(),
        total: total,
      );
    }
    return (items: <OrderModel>[], total: 0);
  }

  Future<OrderModel?> getOrder(String orderNo) async {
    final response = await _api.get('/orders/$orderNo');
    final data = response.data;
    if (data['code'] == 0 && data['data'] != null) {
      return OrderModel.fromJson(data['data'] as Map<String, dynamic>);
    }
    return null;
  }

  Future<Map<String, dynamic>> createOrder(List<({String productId, int quantity})> items) async {
    final body = {
      'items': items.map((e) => {'product_id': e.productId, 'quantity': e.quantity}).toList(),
    };
    final response = await _api.post('/orders', data: body);
    return response.data;
  }

  Future<Map<String, dynamic>> payOrder(String orderNo, {required String paymentMethod, required String paymentId}) async {
    final body = {'payment_method': paymentMethod, 'payment_id': paymentId};
    final response = await _api.post('/orders/$orderNo/pay', data: body);
    return response.data;
  }

  Future<Map<String, dynamic>> cancelOrder(String orderNo) async {
    final response = await _api.post('/orders/$orderNo/cancel');
    return response.data;
  }

  Future<Map<String, dynamic>> shipOrder(String orderNo) async {
    final response = await _api.post('/orders/$orderNo/ship');
    return response.data;
  }

  Future<Map<String, dynamic>> confirmReceipt(String orderNo) async {
    final response = await _api.post('/orders/$orderNo/confirm');
    return response.data;
  }

  Future<Map<String, dynamic>> refundOrder(String orderNo) async {
    final response = await _api.post('/orders/$orderNo/refund');
    return response.data;
  }
}

class CommentRepository {
  final ApiClient _api = ApiClient();

  Future<({List<CommentModel> items, int total})> getComments(String postId, {int page = 1, int pageSize = 20}) async {
    final params = <String, dynamic>{'page': page, 'page_size': pageSize};
    final response = await _api.get('/posts/$postId/comments', queryParameters: params);
    final data = response.data;
    if (data['code'] == 0 && data['data'] != null) {
      final items = data['data']['items'] as List;
      final total = data['data']['total'] as int? ?? 0;
      return (
        items: items.map((e) => CommentModel.fromJson(e as Map<String, dynamic>)).toList(),
        total: total,
      );
    }
    return (items: <CommentModel>[], total: 0);
  }

  Future<CommentModel?> createComment(String postId, String content, {String? parentId}) async {
    final body = <String, dynamic>{'content': content};
    if (parentId != null) body['parent_id'] = parentId;
    final response = await _api.post('/posts/$postId/comments', data: body);
    final data = response.data;
    if (data['code'] == 0 && data['data'] != null) {
      return CommentModel.fromJson(data['data'] as Map<String, dynamic>);
    }
    return null;
  }
}
