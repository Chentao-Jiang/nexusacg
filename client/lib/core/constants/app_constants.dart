class AppConstants {
  static const String apiBaseUrl = String.fromEnvironment('API_BASE_URL', defaultValue: 'https://nexusacg-production.up.railway.app/api/v1');

  static const String zoneCosplay = 'cosplay';
  static const String zonePeripheral = 'peripheral';

  static const int defaultPageSize = 20;
}
