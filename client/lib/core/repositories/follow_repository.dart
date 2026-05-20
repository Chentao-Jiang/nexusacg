import 'package:nexusacg/core/network/api_client.dart';

class FollowRepository {
  final _api = ApiClient();

  Future<bool> follow(String userId) async {
    final res = await _api.post('/users/$userId/follow');
    final d = res.data;
    return d is Map && d['code'] == 0;
  }

  Future<bool> unfollow(String userId) async {
    final res = await _api.delete('/users/$userId/follow');
    final d = res.data;
    return d is Map && d['code'] == 0;
  }

  Future<bool> isFollowing(String userId) async {
    final res = await _api.get('/users/$userId/isfollowing');
    final d = res.data;
    if (d is Map && d['code'] == 0 && d['data'] != null) {
      final data = d['data'] as Map;
      return data['is_following'] == true;
    }
    return false;
  }

  Future<Map<String, dynamic>?> getFollowers(String userId, {int page = 1}) async {
    final res = await _api.get('/users/$userId/followers', queryParameters: {'page': page, 'page_size': 20});
    final d = res.data;
    if (d is Map && d['code'] == 0 && d['data'] != null) {
      return d['data'] as Map<String, dynamic>;
    }
    return null;
  }

  Future<Map<String, dynamic>?> getFollowing(String userId, {int page = 1}) async {
    final res = await _api.get('/users/$userId/following', queryParameters: {'page': page, 'page_size': 20});
    final d = res.data;
    if (d is Map && d['code'] == 0 && d['data'] != null) {
      return d['data'] as Map<String, dynamic>;
    }
    return null;
  }
}
