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
}

class EventRepository {
  final ApiClient _api = ApiClient();

  Future<List<EventModel>> getEvents({
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
      return items.map((e) => EventModel.fromJson(e as Map<String, dynamic>)).toList();
    }
    return [];
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
}
