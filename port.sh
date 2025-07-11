kill $(lsof -i :55000 | grep main | awk '{print $2}' | head -n 1)
echo "air was killed"
