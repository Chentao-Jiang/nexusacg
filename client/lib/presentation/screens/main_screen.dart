import 'package:flutter/material.dart';
import 'package:nexusacg/presentation/screens/home/home_screen.dart';
import 'package:nexusacg/presentation/screens/products/product_screen.dart';
import 'package:nexusacg/presentation/screens/community/community_screen.dart';
import 'package:nexusacg/presentation/screens/events/events_screen.dart';
import 'package:nexusacg/presentation/screens/profile/profile_screen.dart';

class MainScreen extends StatefulWidget {
  const MainScreen({super.key});

  @override
  State<MainScreen> createState() => _MainScreenState();
}

class _MainScreenState extends State<MainScreen> {
  int _currentIndex = 0;

  final List<Widget> _screens = const [
    HomeScreen(),
    ProductScreen(),
    CommunityScreen(),
    EventsScreen(),
    ProfileScreen(),
  ];

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: IndexedStack(index: _currentIndex, children: _screens),
      bottomNavigationBar: NavigationBar(
        selectedIndex: _currentIndex,
        onDestinationSelected: (i) => setState(() => _currentIndex = i),
        destinations: const [
          NavigationDestination(icon: Icon(Icons.home_outlined), selectedIcon: Icon(Icons.home), label: '首页'),
          NavigationDestination(icon: Icon(Icons.shopping_bag_outlined), selectedIcon: Icon(Icons.shopping_bag), label: '商城'),
          NavigationDestination(icon: Icon(Icons.forum_outlined), selectedIcon: Icon(Icons.forum), label: '社区'),
          NavigationDestination(icon: Icon(Icons.map_outlined), selectedIcon: Icon(Icons.map), label: '活动'),
          NavigationDestination(icon: Icon(Icons.person_outlined), selectedIcon: Icon(Icons.person), label: '我的'),
        ],
      ),
    );
  }
}
