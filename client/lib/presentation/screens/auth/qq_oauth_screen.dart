import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:flutter_inappwebview/flutter_inappwebview.dart';
import 'package:nexusacg/core/network/api_client.dart';
import 'package:nexusacg/presentation/blocs/auth/auth_bloc.dart';

class QQOAuthLoginScreen extends StatefulWidget {
  final String authUrl;
  const QQOAuthLoginScreen({super.key, required this.authUrl});

  @override
  State<QQOAuthLoginScreen> createState() => _QQOAuthLoginScreenState();
}

class _QQOAuthLoginScreenState extends State<QQOAuthLoginScreen> {
  bool _loading = false;

  Future<void> _handleCallback(String code) async {
    setState(() => _loading = true);
    try {
      final result = await ApiClient().get('/auth/qq/callback', queryParameters: {'code': code});
      final data = result.data;
      if (data != null && data['code'] == 0 && data['data'] != null) {
        final token = data['data']['access_token'] as String;
        ApiClient().accessToken = token;

        if (mounted) {
          Navigator.of(context).pop();
          context.read<AuthBloc>().add(AuthTokenRestored(token));
        }
      } else {
        if (mounted) {
          Navigator.of(context).pop();
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(content: Text(data?['message'] as String? ?? 'QQ 登录失败')),
          );
        }
      }
    } catch (e) {
      if (mounted) {
        Navigator.of(context).pop();
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('QQ 登录失败: $e')),
        );
      }
    } finally {
      if (mounted) setState(() => _loading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('QQ 登录'),
        leading: IconButton(
          icon: const Icon(Icons.close),
          onPressed: () => Navigator.of(context).pop(),
        ),
      ),
      body: Stack(
        children: [
          InAppWebView(
            initialUrlRequest: URLRequest(url: WebUri(widget.authUrl)),
            initialSettings: InAppWebViewSettings(
              useShouldOverrideUrlLoading: true,
              mediaPlaybackRequiresUserGesture: false,
            ),
            shouldOverrideUrlLoading: (controller, navigationAction) async {
              final uri = navigationAction.request.url;
              if (uri != null &&
                  uri.path.contains('/auth/qq/callback') &&
                  uri.queryParameters['code'] != null) {
                _handleCallback(uri.queryParameters['code']!);
                return NavigationActionPolicy.CANCEL;
              }
              return NavigationActionPolicy.ALLOW;
            },
          ),
          if (_loading)
            Container(
              color: Colors.black26,
              child: const Center(child: CircularProgressIndicator()),
            ),
        ],
      ),
    );
  }
}
