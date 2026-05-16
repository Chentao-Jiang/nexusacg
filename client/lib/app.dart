import 'dart:async';
import 'dart:io';

import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:url_launcher/url_launcher.dart';
import 'package:nexusacg/presentation/blocs/auth/auth_bloc.dart';
import 'package:nexusacg/presentation/blocs/auth/auth_state.dart';
import 'package:nexusacg/presentation/blocs/product/product_bloc.dart';
import 'package:nexusacg/presentation/screens/main_screen.dart';
import 'package:nexusacg/presentation/screens/auth/login_screen.dart';
import 'package:nexusacg/presentation/screens/auth/email_register_screen.dart';
import 'package:nexusacg/presentation/screens/auth/email_pending_screen.dart';
import 'package:nexusacg/presentation/screens/auth/verify_success_screen.dart';
import 'package:nexusacg/core/theme/app_theme.dart';

class NexusACGApp extends StatefulWidget {
  const NexusACGApp({super.key});

  @override
  State<NexusACGApp> createState() => _NexusACGAppState();
}

class _NexusACGAppState extends State<NexusACGApp> {
  late final AuthBloc _authBloc;
  final GlobalKey<NavigatorState> _navigatorKey = GlobalKey<NavigatorState>();
  StreamSubscription<String>? _linkSubscription;

  @override
  void initState() {
    super.initState();
    _authBloc = AuthBloc();
    _restoreAuth();
    _handleInitialLink();
  }

  Future<void> _handleInitialLink() async {
    try {
      if (Platform.isAndroid) {
        // Read the initial route from the platform (populated by Android intent data)
        final route = WidgetsBinding.instance.platformDispatcher.defaultRouteName;
        if (route != null && route != '/' && route.isNotEmpty) {
          _handleDeepLink(route);
        }
      }
    } catch (_) {}
  }

  void _handleDeepLink(String link) {
    final uri = Uri.tryParse(link);
    if (uri == null) return;
    final nav = _navigatorKey.currentState;
    if (nav == null) return;

    if (uri.scheme == 'nexusacg') {
      if (uri.host == 'verify') {
        final token = uri.queryParameters['token'];
        if (token != null) {
          nav.pushNamed('/verify-success', arguments: token);
        }
      } else if (uri.host == 'login') {
        nav.pushNamed('/login');
      }
    } else if (uri.path == '/verify' || uri.host == 'verify') {
      final token = uri.queryParameters['token'];
      if (token != null) {
        nav.pushNamed('/verify-success', arguments: token);
      }
    } else if (uri.path == '/login' || uri.host == 'login') {
      nav.pushNamed('/login');
    }
  }

  Future<void> _restoreAuth() async {
    final prefs = await SharedPreferences.getInstance();
    final token = prefs.getString('access_token');
    if (token != null && token.isNotEmpty) {
      _authBloc.add(AuthTokenRestored(token));
    }
  }

  @override
  Widget build(BuildContext context) {
    return MultiBlocProvider(
      providers: [
        BlocProvider.value(value: _authBloc),
        BlocProvider(create: (_) => ProductBloc()),
      ],
      child: MaterialApp(
        navigatorKey: _navigatorKey,
        title: '次元链 NexusACG',
        debugShowCheckedModeBanner: false,
        theme: AppTheme.lightTheme,
        darkTheme: AppTheme.darkTheme,
        themeMode: ThemeMode.system,
        home: const AppEntryPoint(),
        routes: {
          '/login': (context) => const LoginScreen(),
          '/email-register': (context) => const EmailRegisterScreen(),
          '/verify-success': (context) {
            final token = ModalRoute.of(context)?.settings.arguments as String?;
            return VerifySuccessScreen(nickname: '');
          },
        },
      ),
    );
  }

  @override
  void dispose() {
    _linkSubscription?.cancel();
    _authBloc.close();
    super.dispose();
  }
}

class AppEntryPoint extends StatelessWidget {
  const AppEntryPoint({super.key});

  @override
  Widget build(BuildContext context) {
    return BlocBuilder<AuthBloc, AuthState>(
      builder: (context, state) {
        if (state is AuthAuthenticated) {
          return const MainScreen();
        }
        if (state is AuthLoading) {
          return const Scaffold(
            body: Center(child: CircularProgressIndicator()),
          );
        }
        return const LoginScreen();
      },
    );
  }
}
