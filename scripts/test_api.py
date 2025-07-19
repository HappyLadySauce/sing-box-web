#!/usr/bin/env python3
"""
Comprehensive API Testing Script for sing-box-web
Tests all backend API endpoints with proper authentication and validation.
"""

import requests
import json
import time
import sys
import argparse
from typing import Dict, List, Optional, Any
from datetime import datetime, timedelta
import logging

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

class SingBoxAPITester:
    def __init__(self, base_url: str = "http://localhost:8080", verbose: bool = False):
        self.base_url = base_url.rstrip('/')
        self.api_base = f"{self.base_url}/api/v1"
        self.access_token = None
        self.refresh_token = None
        self.verbose = verbose
        self.session = requests.Session()
        self.test_results = []
        
        # Test data
        self.test_user_id = None
        self.test_node_id = None
        self.test_plan_id = None
        
    def log_test(self, test_name: str, success: bool, details: str = ""):
        """Log test result"""
        status = "PASS" if success else "FAIL"
        self.test_results.append({
            "test": test_name,
            "status": status,
            "details": details,
            "timestamp": datetime.now().isoformat()
        })
        
        if self.verbose or not success:
            logger.info(f"{status}: {test_name} - {details}")

    def make_request(self, method: str, endpoint: str, data: dict = None, 
                    headers: dict = None, auth_required: bool = True) -> requests.Response:
        """Make HTTP request with optional authentication"""
        url = f"{self.api_base}{endpoint}"
        
        # Default headers
        request_headers = {"Content-Type": "application/json"}
        if headers:
            request_headers.update(headers)
            
        # Add auth header if required and available
        if auth_required and self.access_token:
            request_headers["Authorization"] = f"Bearer {self.access_token}"
        
        try:
            if method.upper() == "GET":
                response = self.session.get(url, headers=request_headers, params=data)
            elif method.upper() == "POST":
                response = self.session.post(url, headers=request_headers, 
                                           json=data if data else None)
            elif method.upper() == "PUT":
                response = self.session.put(url, headers=request_headers, 
                                          json=data if data else None)
            elif method.upper() == "DELETE":
                response = self.session.delete(url, headers=request_headers)
            else:
                raise ValueError(f"Unsupported HTTP method: {method}")
                
            return response
        except requests.exceptions.RequestException as e:
            logger.error(f"Request failed: {e}")
            return None

    def test_health_endpoints(self):
        """Test health check endpoints"""
        logger.info("Testing health endpoints...")
        
        # Test /health
        try:
            response = requests.get(f"{self.base_url}/health")
            if response.status_code == 200:
                self.log_test("Health Check", True, f"Status: {response.status_code}")
            else:
                self.log_test("Health Check", False, f"Status: {response.status_code}")
        except Exception as e:
            self.log_test("Health Check", False, str(e))
            
        # Test /livez
        try:
            response = requests.get(f"{self.base_url}/livez")
            if response.status_code == 200:
                self.log_test("Liveness Check", True, f"Status: {response.status_code}")
            else:
                self.log_test("Liveness Check", False, f"Status: {response.status_code}")
        except Exception as e:
            self.log_test("Liveness Check", False, str(e))
            
        # Test /readyz
        try:
            response = requests.get(f"{self.base_url}/readyz")
            if response.status_code == 200:
                self.log_test("Readiness Check", True, f"Status: {response.status_code}")
            else:
                self.log_test("Readiness Check", False, f"Status: {response.status_code}")
        except Exception as e:
            self.log_test("Readiness Check", False, str(e))

    def test_authentication(self):
        """Test authentication endpoints"""
        logger.info("Testing authentication endpoints...")
        
        # Test registration
        register_data = {
            "username": f"testuser_{int(time.time())}",
            "email": f"test_{int(time.time())}@example.com",
            "password": "testpassword123",
            "display_name": "Test User"
        }
        
        response = self.make_request("POST", "/auth/register", register_data, auth_required=False)
        if response and response.status_code == 201:
            self.log_test("User Registration", True, f"Created user: {register_data['username']}")
            try:
                data = response.json()
                if 'data' in data and 'access_token' in data['data']:
                    self.access_token = data['data']['access_token']
                    self.refresh_token = data['data']['refresh_token']
            except:
                pass
        else:
            status_code = response.status_code if response else "No response"
            self.log_test("User Registration", False, f"Status: {status_code}")
        
        # Test login with admin credentials
        login_data = {
            "username": "admin",
            "password": "admin123"
        }
        
        response = self.make_request("POST", "/auth/login", login_data, auth_required=False)
        if response and response.status_code == 200:
            self.log_test("Admin Login", True, "Successfully logged in as admin")
            try:
                data = response.json()
                self.access_token = data['access_token']
                self.refresh_token = data['refresh_token']
            except:
                self.log_test("Admin Login", False, "Failed to parse tokens")
        else:
            status_code = response.status_code if response else "No response"
            self.log_test("Admin Login", False, f"Status: {status_code}")
            
        # Test profile endpoint
        response = self.make_request("GET", "/auth/profile")
        if response and response.status_code == 200:
            self.log_test("Get Profile", True, "Retrieved user profile")
        else:
            status_code = response.status_code if response else "No response"
            self.log_test("Get Profile", False, f"Status: {status_code}")
            
        # Test token refresh
        if self.refresh_token:
            refresh_data = {"refresh_token": self.refresh_token}
            response = self.make_request("POST", "/auth/refresh", refresh_data, auth_required=False)
            if response and response.status_code == 200:
                self.log_test("Token Refresh", True, "Successfully refreshed token")
                try:
                    data = response.json()
                    self.access_token = data['access_token']
                except:
                    pass
            else:
                status_code = response.status_code if response else "No response"
                self.log_test("Token Refresh", False, f"Status: {status_code}")

    def test_user_management(self):
        """Test user management endpoints"""
        logger.info("Testing user management endpoints...")
        
        # List users
        response = self.make_request("GET", "/admin/users")
        if response and response.status_code == 200:
            self.log_test("List Users", True, f"Retrieved users list")
            try:
                data = response.json()
                if 'data' in data and len(data['data']) > 0:
                    self.test_user_id = data['data'][0]['id']
            except:
                pass
        else:
            status_code = response.status_code if response else "No response"
            self.log_test("List Users", False, f"Status: {status_code}")
            
        # Create user
        user_data = {
            "username": f"apitest_{int(time.time())}",
            "email": f"apitest_{int(time.time())}@example.com",
            "password": "testpassword123",
            "display_name": "API Test User",
            "plan_id": 1,
            "traffic_quota": 10737418240,  # 10GB
            "device_limit": 3,
            "speed_limit": 0
        }
        
        response = self.make_request("POST", "/admin/users", user_data)
        if response and response.status_code == 201:
            self.log_test("Create User", True, f"Created user: {user_data['username']}")
            try:
                data = response.json()
                if 'data' in data and 'id' in data['data']:
                    self.test_user_id = data['data']['id']
            except:
                pass
        else:
            status_code = response.status_code if response else "No response"
            self.log_test("Create User", False, f"Status: {status_code}")
            
        # Get user details
        if self.test_user_id:
            response = self.make_request("GET", f"/admin/users/{self.test_user_id}")
            if response and response.status_code == 200:
                self.log_test("Get User", True, f"Retrieved user {self.test_user_id}")
            else:
                status_code = response.status_code if response else "No response"
                self.log_test("Get User", False, f"Status: {status_code}")
                
            # Update user
            update_data = {
                "display_name": "Updated API Test User",
                "device_limit": 5
            }
            response = self.make_request("PUT", f"/admin/users/{self.test_user_id}", update_data)
            if response and response.status_code == 200:
                self.log_test("Update User", True, f"Updated user {self.test_user_id}")
            else:
                status_code = response.status_code if response else "No response"
                self.log_test("Update User", False, f"Status: {status_code}")
                
            # Reset user traffic
            response = self.make_request("POST", f"/admin/users/{self.test_user_id}/reset-traffic")
            if response and response.status_code == 200:
                self.log_test("Reset User Traffic", True, f"Reset traffic for user {self.test_user_id}")
            else:
                status_code = response.status_code if response else "No response"
                self.log_test("Reset User Traffic", False, f"Status: {status_code}")

    def test_node_management(self):
        """Test node management endpoints"""
        logger.info("Testing node management endpoints...")
        
        # List nodes
        response = self.make_request("GET", "/admin/nodes")
        if response and response.status_code == 200:
            self.log_test("List Nodes", True, "Retrieved nodes list")
            try:
                data = response.json()
                if 'data' in data and len(data['data']) > 0:
                    self.test_node_id = data['data'][0]['id']
            except:
                pass
        else:
            status_code = response.status_code if response else "No response"
            self.log_test("List Nodes", False, f"Status: {status_code}")
            
        # Create node
        node_data = {
            "name": f"Test Node {int(time.time())}",
            "description": "API Test Node",
            "type": "vmess",
            "host": "test.example.com",
            "port": 443,
            "uuid": "550e8400-e29b-41d4-a716-446655440000",
            "region": "US",
            "country": "United States",
            "city": "New York",
            "max_users": 100,
            "speed_limit": 0,
            "traffic_rate": 1.0,
            "is_enabled": True
        }
        
        response = self.make_request("POST", "/admin/nodes", node_data)
        if response and response.status_code == 201:
            self.log_test("Create Node", True, f"Created node: {node_data['name']}")
            try:
                data = response.json()
                if 'data' in data and 'id' in data['data']:
                    self.test_node_id = data['data']['id']
            except:
                pass
        else:
            status_code = response.status_code if response else "No response"
            self.log_test("Create Node", False, f"Status: {status_code}")
            
        # Get node details
        if self.test_node_id:
            response = self.make_request("GET", f"/admin/nodes/{self.test_node_id}")
            if response and response.status_code == 200:
                self.log_test("Get Node", True, f"Retrieved node {self.test_node_id}")
            else:
                status_code = response.status_code if response else "No response"
                self.log_test("Get Node", False, f"Status: {status_code}")
                
            # Update node
            update_data = {
                "description": "Updated API Test Node",
                "max_users": 200
            }
            response = self.make_request("PUT", f"/admin/nodes/{self.test_node_id}", update_data)
            if response and response.status_code == 200:
                self.log_test("Update Node", True, f"Updated node {self.test_node_id}")
            else:
                status_code = response.status_code if response else "No response"
                self.log_test("Update Node", False, f"Status: {status_code}")

    def test_plan_management(self):
        """Test plan management endpoints"""
        logger.info("Testing plan management endpoints...")
        
        # List plans
        response = self.make_request("GET", "/admin/plans")
        if response and response.status_code == 200:
            self.log_test("List Plans", True, "Retrieved plans list")
            try:
                data = response.json()
                if 'data' in data and len(data['data']) > 0:
                    self.test_plan_id = data['data'][0]['id']
            except:
                pass
        else:
            status_code = response.status_code if response else "No response"
            self.log_test("List Plans", False, f"Status: {status_code}")
            
        # Create plan
        plan_data = {
            "name": f"Test Plan {int(time.time())}",
            "description": "API Test Plan",
            "period": "monthly",
            "price": 1000,  # $10.00
            "currency": "USD",
            "traffic_quota": 107374182400,  # 100GB
            "speed_limit": 0,
            "device_limit": 5,
            "connection_limit": 0,
            "allowed_protocols": ["vmess", "vless", "trojan"],
            "allowed_nodes": [],
            "enable_file_sharing": True,
            "enable_port_forwarding": False,
            "enable_p2p": True,
            "enable_torrent": False,
            "priority": 1,
            "bandwidth_ratio": 1.0,
            "restriction_level": 0,
            "blocked_domains": [],
            "allowed_countries": ["US", "EU", "AS"],
            "is_trial_plan": False,
            "trial_days": 0,
            "is_promotional": False,
            "promotion_price": 0,
            "is_public": True,
            "is_enabled": True,
            "max_users": 0,
            "color": "#007bff",
            "icon": "star",
            "sort_order": 1,
            "is_recommended": False,
            "features": {
                "priority_support": True,
                "custom_dns": True
            },
            "metadata": {
                "created_by": "api_test"
            }
        }
        
        response = self.make_request("POST", "/admin/plans", plan_data)
        if response and response.status_code == 201:
            self.log_test("Create Plan", True, f"Created plan: {plan_data['name']}")
            try:
                data = response.json()
                if 'data' in data and 'id' in data['data']:
                    self.test_plan_id = data['data']['id']
            except:
                pass
        else:
            status_code = response.status_code if response else "No response"
            self.log_test("Create Plan", False, f"Status: {status_code}")
            
        # Get plan details
        if self.test_plan_id:
            response = self.make_request("GET", f"/admin/plans/{self.test_plan_id}")
            if response and response.status_code == 200:
                self.log_test("Get Plan", True, f"Retrieved plan {self.test_plan_id}")
            else:
                status_code = response.status_code if response else "No response"
                self.log_test("Get Plan", False, f"Status: {status_code}")

    def test_traffic_statistics(self):
        """Test traffic statistics endpoints"""
        logger.info("Testing traffic statistics endpoints...")
        
        # Get traffic statistics
        end_date = datetime.now()
        start_date = end_date - timedelta(days=7)
        
        params = {
            "start_date": start_date.strftime("%Y-%m-%d"),
            "end_date": end_date.strftime("%Y-%m-%d"),
            "granularity": "daily",
            "include_top": "true"
        }
        
        response = self.make_request("GET", "/traffic/statistics", params)
        if response and response.status_code == 200:
            self.log_test("Traffic Statistics", True, "Retrieved traffic statistics")
        else:
            status_code = response.status_code if response else "No response"
            self.log_test("Traffic Statistics", False, f"Status: {status_code}")
            
        # Get traffic chart
        response = self.make_request("GET", "/traffic/chart", params)
        if response and response.status_code == 200:
            self.log_test("Traffic Chart", True, "Retrieved traffic chart data")
        else:
            status_code = response.status_code if response else "No response"
            self.log_test("Traffic Chart", False, f"Status: {status_code}")
            
        # Get live traffic
        response = self.make_request("GET", "/traffic/live")
        if response and response.status_code == 200:
            self.log_test("Live Traffic", True, "Retrieved live traffic data")
        else:
            status_code = response.status_code if response else "No response"
            self.log_test("Live Traffic", False, f"Status: {status_code}")
            
        # Get traffic summary
        response = self.make_request("GET", "/traffic/summary")
        if response and response.status_code == 200:
            self.log_test("Traffic Summary", True, "Retrieved traffic summary")
        else:
            status_code = response.status_code if response else "No response"
            self.log_test("Traffic Summary", False, f"Status: {status_code}")
            
        # Get top users
        response = self.make_request("GET", "/traffic/top-users", {"limit": 10})
        if response and response.status_code == 200:
            self.log_test("Top Users Traffic", True, "Retrieved top users")
        else:
            status_code = response.status_code if response else "No response"
            self.log_test("Top Users Traffic", False, f"Status: {status_code}")
            
        # Get top nodes
        response = self.make_request("GET", "/traffic/top-nodes", {"limit": 10})
        if response and response.status_code == 200:
            self.log_test("Top Nodes Traffic", True, "Retrieved top nodes")
        else:
            status_code = response.status_code if response else "No response"
            self.log_test("Top Nodes Traffic", False, f"Status: {status_code}")

    def test_system_endpoints(self):
        """Test system monitoring endpoints"""
        logger.info("Testing system endpoints...")
        
        # System status
        response = self.make_request("GET", "/system/status")
        if response and response.status_code == 200:
            self.log_test("System Status", True, "Retrieved system status")
        else:
            status_code = response.status_code if response else "No response"
            self.log_test("System Status", False, f"Status: {status_code}")
            
        # Dashboard data
        response = self.make_request("GET", "/system/dashboard")
        if response and response.status_code == 200:
            self.log_test("System Dashboard", True, "Retrieved dashboard data")
        else:
            status_code = response.status_code if response else "No response"
            self.log_test("System Dashboard", False, f"Status: {status_code}")
            
        # System statistics
        response = self.make_request("GET", "/system/statistics")
        if response and response.status_code == 200:
            self.log_test("System Statistics", True, "Retrieved system statistics")
        else:
            status_code = response.status_code if response else "No response"
            self.log_test("System Statistics", False, f"Status: {status_code}")

    def test_user_node_relationships(self):
        """Test user-node relationship endpoints"""
        logger.info("Testing user-node relationships...")
        
        if self.test_user_id and self.test_node_id:
            # Add user to node
            response = self.make_request("POST", f"/admin/users/{self.test_user_id}/nodes/{self.test_node_id}")
            if response and response.status_code == 200:
                self.log_test("Add User to Node", True, f"Added user {self.test_user_id} to node {self.test_node_id}")
            else:
                status_code = response.status_code if response else "No response"
                self.log_test("Add User to Node", False, f"Status: {status_code}")
                
            # Get user nodes
            response = self.make_request("GET", f"/admin/users/{self.test_user_id}/nodes")
            if response and response.status_code == 200:
                self.log_test("Get User Nodes", True, f"Retrieved nodes for user {self.test_user_id}")
            else:
                status_code = response.status_code if response else "No response"
                self.log_test("Get User Nodes", False, f"Status: {status_code}")
                
            # Get node users
            response = self.make_request("GET", f"/admin/nodes/{self.test_node_id}/users")
            if response and response.status_code == 200:
                self.log_test("Get Node Users", True, f"Retrieved users for node {self.test_node_id}")
            else:
                status_code = response.status_code if response else "No response"
                self.log_test("Get Node Users", False, f"Status: {status_code}")

    def cleanup_test_data(self):
        """Clean up test data"""
        logger.info("Cleaning up test data...")
        
        # Delete test user
        if self.test_user_id:
            response = self.make_request("DELETE", f"/admin/users/{self.test_user_id}")
            if response and response.status_code == 200:
                self.log_test("Cleanup: Delete User", True, f"Deleted user {self.test_user_id}")
            else:
                status_code = response.status_code if response else "No response"
                self.log_test("Cleanup: Delete User", False, f"Status: {status_code}")
                
        # Delete test node
        if self.test_node_id:
            response = self.make_request("DELETE", f"/admin/nodes/{self.test_node_id}")
            if response and response.status_code == 200:
                self.log_test("Cleanup: Delete Node", True, f"Deleted node {self.test_node_id}")
            else:
                status_code = response.status_code if response else "No response"
                self.log_test("Cleanup: Delete Node", False, f"Status: {status_code}")
                
        # Delete test plan
        if self.test_plan_id:
            response = self.make_request("DELETE", f"/admin/plans/{self.test_plan_id}")
            if response and response.status_code == 200:
                self.log_test("Cleanup: Delete Plan", True, f"Deleted plan {self.test_plan_id}")
            else:
                status_code = response.status_code if response else "No response"
                self.log_test("Cleanup: Delete Plan", False, f"Status: {status_code}")

    def run_all_tests(self):
        """Run all API tests"""
        logger.info(f"Starting API tests for {self.base_url}")
        start_time = time.time()
        
        try:
            # Test sequence
            self.test_health_endpoints()
            self.test_authentication()
            self.test_user_management()
            self.test_node_management()
            self.test_plan_management()
            self.test_traffic_statistics()
            self.test_system_endpoints()
            self.test_user_node_relationships()
            
        finally:
            # Always cleanup
            self.cleanup_test_data()
            
        end_time = time.time()
        duration = end_time - start_time
        
        # Print summary
        self.print_summary(duration)

    def print_summary(self, duration: float):
        """Print test summary"""
        total_tests = len(self.test_results)
        passed_tests = len([r for r in self.test_results if r['status'] == 'PASS'])
        failed_tests = total_tests - passed_tests
        
        print("\n" + "="*60)
        print("API TEST SUMMARY")
        print("="*60)
        print(f"Total Tests: {total_tests}")
        print(f"Passed: {passed_tests}")
        print(f"Failed: {failed_tests}")
        print(f"Success Rate: {(passed_tests/total_tests)*100:.1f}%")
        print(f"Duration: {duration:.2f} seconds")
        print("="*60)
        
        if failed_tests > 0:
            print("\nFAILED TESTS:")
            print("-"*40)
            for result in self.test_results:
                if result['status'] == 'FAIL':
                    print(f"❌ {result['test']}: {result['details']}")
        
        if self.verbose:
            print("\nDETAILED RESULTS:")
            print("-"*40)
            for result in self.test_results:
                status_emoji = "✅" if result['status'] == 'PASS' else "❌"
                print(f"{status_emoji} {result['test']}: {result['details']}")
                
        # Save results to file
        timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
        filename = f"api_test_results_{timestamp}.json"
        with open(filename, 'w') as f:
            json.dump({
                "summary": {
                    "total_tests": total_tests,
                    "passed_tests": passed_tests,
                    "failed_tests": failed_tests,
                    "success_rate": (passed_tests/total_tests)*100,
                    "duration": duration,
                    "timestamp": datetime.now().isoformat()
                },
                "results": self.test_results
            }, f, indent=2)
        
        print(f"\nDetailed results saved to: {filename}")

def main():
    parser = argparse.ArgumentParser(description='Test sing-box-web API endpoints')
    parser.add_argument('--url', default='http://localhost:8080', 
                       help='Base URL of the API server (default: http://localhost:8080)')
    parser.add_argument('--verbose', '-v', action='store_true',
                       help='Enable verbose output')
    
    args = parser.parse_args()
    
    tester = SingBoxAPITester(base_url=args.url, verbose=args.verbose)
    tester.run_all_tests()

if __name__ == "__main__":
    main()